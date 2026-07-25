
import { useState, useEffect, useRef } from 'react';
import './topic_menu.css';
import { GetTopic, SubmitRssUrl, UnlinkRssFromTopic } from '../../wailsjs/go/main/App';
import type { dto } from '../../wailsjs/go/models';

interface TopicMenuProps {
    topicId: number;
    topic: string; // display name
}

function TopicMenu({ topicId, topic }: TopicMenuProps) {
    const [isDropDownOpen, setIsDropDownOpen] = useState(false);
    const [rssList, setRssList] = useState<dto.RssItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string>('');
    const hasFetched = useRef(false);

    const [isAddingRss, setIsAddingRss] = useState(false);
    const [rssUrl, setRssUrl] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const [unlinkingId, setUnlinkingId] = useState<number | null>(null);
    const loadTopic = () => {
        setLoading(true);
        setError('');
        GetTopic(topicId)
            .then((topicResp: dto.TopicResponse) => {
                setRssList(topicResp?.rss ?? []);
                hasFetched.current = true;
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to fetch topic');
            })
            .finally(() => setLoading(false));
    };

    useEffect(() => {
        if (!isDropDownOpen || hasFetched.current) return;
        loadTopic();
    }, [isDropDownOpen, topicId]);

    const handleAddRss = (e: React.FormEvent) => {
        e.preventDefault();
        const url = rssUrl.trim();
        if (!url || submitting) return;

        setSubmitting(true);
        setError('');
        SubmitRssUrl(url, topicId)
            .then((item: dto.RssItem) => {
                setRssList((prev) => [item, ...prev]);
                setRssUrl('');
                setIsAddingRss(false);
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to add feed');
            })
            .finally(() => setSubmitting(false));
    };

    const handleUnlink = (rssId: number, rssTitle: string) => {
        const confirmed = window.confirm(
            `Remove "${rssTitle}" from "${topic}"? The feed itself won't be deleted, ` +
            `only its link to this topic. You can find it again if it's linked elsewhere.`
        );
        if (!confirmed) return;

        setUnlinkingId(rssId);
        setError('');
        UnlinkRssFromTopic(rssId, topicId)
            .then(() => {
                setRssList((prev) => prev.filter((r) => r.id !== rssId));
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to remove feed');
            })
            .finally(() => setUnlinkingId(null));
    };

    return (
        <div className="dropdown">
            <button
                className="dropdown-toggle"
                onClick={() => setIsDropDownOpen(!isDropDownOpen)}
            >
                <span className="topic-text">{topic}</span>
                <span className={`arrow ${isDropDownOpen ? 'up' : 'down'}`}>▼</span>
            </button>

            {isDropDownOpen && (
                <div className="dropdown-panel">
                    {error && <div className="dropdown-error">{error}</div>}

                    <ul className="dropdown-menu-list">
                        {loading && <li className="dropdown-item">Loading...</li>}

                        {!loading && rssList.length === 0 && (
                            <li className="dropdown-item">No subscriptions yet</li>
                        )}

                        {!loading && rssList.map((rss) => (
                            <li key={rss.id} className="dropdown-item rss-item">
                                <span>{rss.title}</span>
                                <button
                                    className="rss-unlink-btn"
                                    onClick={() => handleUnlink(rss.id, rss.title)}
                                    disabled={unlinkingId === rss.id}
                                    aria-label={`Remove ${rss.title} from ${topic}`}
                                >
                                    {unlinkingId === rss.id ? '...' : '✕'}
                                </button>
                            </li>
                        ))}
                    </ul>

                    {isAddingRss ? (
                        <form className="rss-add-form" onSubmit={handleAddRss}>
                            <input
                                type="url"
                                autoFocus
                                value={rssUrl}
                                onChange={(e) => setRssUrl(e.target.value)}
                                placeholder="https://example.com/feed.xml"
                                disabled={submitting}
                                onKeyDown={(e) => {
                                    if (e.key === 'Escape') {
                                        setIsAddingRss(false);
                                        setRssUrl('');
                                    }
                                }}
                            />
                            <div className="rss-add-actions">
                                <button type="submit" disabled={submitting || !rssUrl.trim()}>
                                    {submitting ? 'Adding...' : 'Add'}
                                </button>
                                <button
                                    type="button"
                                    onClick={() => { setIsAddingRss(false); setRssUrl(''); }}
                                    disabled={submitting}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    ) : (
                        <button className="rss-add-btn" onClick={() => setIsAddingRss(true)}>
                            + Add RSS
                        </button>
                    )}
                </div>
            )}
        </div>
    );}

export default TopicMenu;

