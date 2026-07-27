import { useEffect, useState } from 'react';
import { GetRssDetail } from '../../../wailsjs/go/main/App';
import type { dto } from '../../../wailsjs/go/models';

interface RssContentProps {
    rssId: number;
    rssTitle: string;
    onSelectPost: (postId: number, postTitle: string) => void;
}

function RssContent({ rssId, rssTitle, onSelectPost }: RssContentProps) {
    const [detail, setDetail] = useState<dto.RssResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

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

    if (loading) return <div className="content-loading">Loading {rssTitle}…</div>;
    if (error) return <div className="content-error">{error}</div>;

    const posts = detail?.posts ?? [];

    return (
        <div className="rss-content">
            <h2>{rssTitle}</h2>
            {posts.length === 0 && <p>No posts yet.</p>}
<ul className="post-list">
    {posts.map((p: dto.PostItem) => (
        <li key={p.id} className="post-list-item">
            <button className="post-list-title" onClick={() => onSelectPost(p.id, p.title)}>
                {p.title}
            </button>
            <span className="post-list-date">{p.publishedAt}</span>
        </li>
    ))}
</ul>
        </div>
    );
}

export default RssContent;
