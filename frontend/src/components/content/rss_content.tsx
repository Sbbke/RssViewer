import { useEffect, useRef, useState } from 'react';
import { GetRssDetail } from '../../../wailsjs/go/main/App';
import type { dto } from '../../../wailsjs/go/models';
import "./rss_content.css"

const PAGE_SIZE = 5;

interface RssContentProps {
    rssId: number;
    rssTitle: string;
    onSelectPost: (postId: number, postTitle: string) => void;
}
function RssContent({ rssId, rssTitle, onSelectPost }: RssContentProps) {
    const [detail, setDetail] = useState<dto.RssResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        let cancelled = false;
        setLoading(true);
        setError('');
        GetRssDetail(rssId)
            .then((data: dto.RssResponse) => {
                if (!cancelled) setDetail(data);
            })
            .catch((err) => {
                if (!cancelled) setError(typeof err === 'string' ? err : err?.message || 'Failed to load feed');
            })
            .finally(() => {
                if (!cancelled) setLoading(false);
            });
        return () => { cancelled = true; };
    }, [rssId]);

    // Reset pagination whenever the feed changes
    useEffect(() => {
        setVisibleCount(PAGE_SIZE);
        if (scrollRef.current) scrollRef.current.scrollTop = 0;
    }, [rssId]);

    const posts = detail?.posts ?? [];

    const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
        const el = e.currentTarget;
        const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
        if (nearBottom) {
            setVisibleCount((prev) => Math.min(prev + PAGE_SIZE, posts.length));
        }
    };

    if (loading) return <div className="content-loading">Loading {rssTitle}…</div>;
    if (error) return <div className="content-error">{error}</div>;

    const visiblePosts = posts.slice(0, visibleCount);

    return (
        <div className="rss-content">
            <h2>{rssTitle}</h2>
            {posts.length === 0 && <p>No posts yet.</p>}
            {posts.length > 0 && (
                <div
                    className="post-list-scroll scroll-panel"
                    ref={scrollRef}
                    onScroll={handleScroll}
                >
                    <ul className="post-list">
                        {visiblePosts.map((p: dto.PostItem) => (
                            <li key={p.id} className="post-list-item">
                                <button
                                    className="post-list-title"
                                    onClick={() => onSelectPost(p.id, p.title)}
                                >
                                    {p.title}
                                </button>
                                <div className="post-list-meta">
                                    <span className="post-list-date">
                                        {new Date(p.publishedAt).toLocaleDateString()}
                                    </span>
                                </div>
                            </li>
                        ))}
                    </ul>
                    {visibleCount < posts.length && (
                        <div className="post-list-loading-more">Loading more…</div>
                    )}
                </div>
            )}
        </div>
    );
}
export default RssContent;
