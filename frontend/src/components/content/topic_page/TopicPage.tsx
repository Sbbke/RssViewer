import { useMemo } from 'react';
import type { dto } from '../../../../wailsjs/go/models';
import TopicHeader from './TopicHeader';
import BriefingCarousel from './BriefingCarousel';
import PostsSection, { type TopicPost } from './PostsSection';
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
}

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

function TopicPage({ topicName, topicDescription, data, onBack, onOpenPost }: TopicPageProps) {
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
 

            <div className="topic-page-inner">
                <div className="topic-page-header-row">
                    <button className="back-btn" onClick={onBack} aria-label="Back to main page">
                            ← Back
                    </button>

                    <button className="refresh-btn" onClick={() => {/* handle refresh */}}>
                        Refresh
                    </button>
                </div>
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
                        // Need the owning rssId too (PostDetailResponse doesn't
                        // carry it) — look it up from the feed it came from.
                        const feed = data.rss.find((f) => f.posts.some((p) => p.id === postId));
                        if (feed) onOpenPost(postId, feed.info.id);
                    }}
                />
            </div>
        </div>
    );
}

export default TopicPage;
