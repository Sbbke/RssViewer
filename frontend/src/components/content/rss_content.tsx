import { useCallback, useEffect, useRef, useState } from 'react';
import { GetRssDetail, CheckRssUpdate} from '../../../wailsjs/go/main/App';
import type { dto } from '../../../wailsjs/go/models';
import ErrorModal from '../common/ErrorModal';
import "./rss_content.css"

const PAGE_SIZE = 5;

interface RssContentProps {
    rssId: number;
    rssTitle: string;
    onSelectPost: (postId: number, postTitle: string) => void;
    onBack: () => void;
}
function RssContent({ rssId, rssTitle, onSelectPost, onBack }: RssContentProps) {
    const [detail, setDetail] = useState<dto.RssResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [checking, setChecking] = useState(false);
    const [checkError, setCheckError] = useState('');
    const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);
    const scrollRef = useRef<HTMLDivElement>(null);

    const fetchDetail = useCallback((showSpinner: boolean) => {
        let cancelled = false;
        if (showSpinner) setLoading(true);
        setError('');
        GetRssDetail(rssId)
            .then((data: dto.RssResponse) => {
                if (!cancelled) setDetail(data);
            })
            .catch((err) => {
                if (!cancelled) setError(typeof err === 'string' ? err : err?.message || 'Failed to load feed');
            })
            .finally(() => {
                if (!cancelled && showSpinner) setLoading(false);
            });
        return () => { cancelled = true; };
    }, [rssId]);

    useEffect(() => {
        const cancel = fetchDetail(true);
        return cancel;
    }, [fetchDetail]);

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

    // Backend confirms success/failure via resolve/reject (nil vs
    // non-nil error on the Go side). A resolved promise means it's
    // safe to re-pull — it does NOT guarantee new posts, just that
    // the refresh completed, so we always re-fetch to stay in sync.
    const handleCheckUpdate = () => {
        setChecking(true);
        setCheckError('');
        CheckRssUpdate(rssId)
            .then(() => {
                fetchDetail(false);
            })
            .catch((err) => {
                setCheckError(typeof err === 'string' ? err : err?.message || 'Failed to check for updates');
            })
            .finally(() => {
                setChecking(false);
            });
    };

    if (loading) return <div className="content-loading">Loading {rssTitle}…</div>;

    // Nothing to show behind the popup if the initial load failed —
    // OK sends the user back rather than leaving them on a dead page.
    if (error) {
        return <ErrorModal title="Couldn't load feed" message={error} onClose={onBack} />;
    }

    const visiblePosts = posts.slice(0, visibleCount);

    return (
        <div className="rss-content">
            <div className="rss-content-header">
                <button className="back-btn" onClick={onBack} aria-label="Back to main page">
                    ← Back
                </button>
                <button
                    className="refresh-btn"
                    onClick={handleCheckUpdate}
                    disabled={checking}
                    aria-label="Check for new posts"
                >
                    {checking ? (
                        <>
                            <span className="spinner-tiny" aria-hidden="true" /> Checking…
                        </>
                    ) : (
                        '⟳ Check for updates'
                    )}
                </button>
            </div>
             <h2>{rssTitle}</h2>

            <p className="rss-content-meta">
                {posts.length} {posts.length === 1 ? 'post' : 'posts'}
            </p>
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
            {checkError && (
                <ErrorModal
                    title="Couldn't check for updates"
                    message={checkError}
                    onClose={() => setCheckError('')}
                />
            )}
        </div>
    );
}
export default RssContent;
