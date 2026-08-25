import { useMemo, useRef, useState } from 'react';
import type { dto } from '../../../../wailsjs/go/models';
import TopicHeader from './TopicHeader';
import BriefingCarousel from './BriefingCarousel';
import PostsSection, { type TopicPost } from './PostsSection';
import TopicContent from '../topic_content'; // legacy flat feed-list view — reused as the flip's back face
import "./topic_page.css";

interface TopicPageProps {
    // Not present on TopicAllInOne — the caller already has this from
    // wherever the topic was selected (e.g. the sidebar's topic list),
    // so it's passed in separately rather than guessed at here.
    topicName: string;
    topicDescription?: string;
    data: dto.TopicAllInOne;
    onBack: () => void;
    onOpenPost: (postId: number, rssId: number) => void;
    onRefresh: () => void;
    isRefreshing: boolean;
}

const FLIP_LEG_MS = 300; // each half of the round trip (out, then back)

function excerptFrom(content: string, maxLen = 160): string {
    const stripped = content.replace(/<[^>]+>/g, '').trim();
    return stripped.length > maxLen ? stripped.slice(0, maxLen).trimEnd() + '…' : stripped;
}

function relativeTimeFrom(iso: string): string {
    const diffMs = Date.now() - new Date(iso).getTime();
    const mins = Math.floor(diffMs / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
}

function TopicPage({ topicName, topicDescription, data, onBack, onOpenPost, onRefresh, isRefreshing }: TopicPageProps) {
    // Rotates to 90° (edge-on, invisible), swaps the actual content,
    // then rotates back to 0° — rather than continuing on to 180°.
    // Parking at 180° would leave the element mirrored (rotateY(180)
    // flips its contents horizontally), which is exactly what made
    // the back face look flipped left/right. Going out-and-back
    // through 90° instead means it's never resting at an angle that
    // mirrors anything.
    const [rotation, setRotation] = useState(0);
    const [showLegacy, setShowLegacy] = useState(false);
    const swapTimeout = useRef<ReturnType<typeof setTimeout>>();

    const handleFlip = () => {
        setRotation(90);
        clearTimeout(swapTimeout.current);
        swapTimeout.current = setTimeout(() => {
            setShowLegacy((v) => !v);
            setRotation(0);
        }, FLIP_LEG_MS);
    };

    // Flatten every RSS feed's posts into one list, tagging each with
    // its source feed title (PostDetailResponse itself carries no
    // source/feed reference, so it's attached here from the parent
    // RssDetailResponse while flattening).
    const posts: TopicPost[] = useMemo(() => {
        return data.rss.flatMap((feed) =>
            feed.posts.map((p) => ({
                id: p.id,
                title: p.title,
                excerpt: excerptFrom(p.content),
                source: feed.info.title,
                publishedAt: p.publishedAt,
                // No thumbnail field exists on PostDetailResponse yet.
            }))
        );
    }, [data]);

    // ASSUMPTION: slide bytes are PNG. Confirm with whatever generates
    // BriefingSlideResponse.Slides — swap the mime type below if it's
    // actually JPEG or something else.
    const slideSrcs = (data.slide?.slides ?? []).map((b64) => `data:image/png;base64,${b64}`);

    return (
        <div className="topic-page">
            <div className="topic-page-back-row">
                <button className="back-btn" onClick={onBack} aria-label="Back to main page">
                    ← Back
                </button>
                <button
                    className="flip-btn"
                    onClick={handleFlip}
                    aria-label={showLegacy ? 'Flip to briefing view' : 'Flip to feed list view'}
                    title={showLegacy ? 'Flip to briefing view' : 'Flip to feed list view'}
                >
                    ⟳
                </button>
            </div>

            <div className="topic-page-flip-stage">
                <div
                    className="topic-page-flip-content"
                    style={{ transform: `rotateY(${rotation}deg)` }}
                >
                    {showLegacy ? (
                        <TopicContent topicId={data.topicId} topicName={topicName} />
                    ) : (
                        <div className="topic-page-inner">
                            <TopicHeader
                                title={topicName}
                                description={topicDescription}
                                postCount={posts.length}
                                lastUpdatedLabel={relativeTimeFrom(data.createdAt)}
                            />

                            <BriefingCarousel slides={slideSrcs} />

                            <PostsSection
                                posts={posts}
                                onOpenPost={(postId) => {
                                    // Need the owning rssId too (PostDetailResponse
                                    // doesn't carry it) — look it up from the feed
                                    // it came from.
                                    const feed = data.rss.find((f) => f.posts.some((p) => p.id === postId));
                                    if (feed) onOpenPost(postId, feed.info.id);
                                }}
                                onRefresh={onRefresh}
                                isRefreshing={isRefreshing}
                            />
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default TopicPage;
