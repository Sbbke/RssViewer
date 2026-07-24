
import { useState, useEffect, useRef } from 'react';
import './topic_menu.css';
import { GetTopic } from '../../wailsjs/go/main/App';
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

    useEffect(() => {
        if (!isDropDownOpen || hasFetched.current) return;

        let cancelled = false;
        setLoading(true);
        setError('');

        // Wails returns a standard Promise<dto.TopicResponse>
        GetTopic(topicId)
            .then((topicResp: dto.TopicResponse) => {
                if (cancelled) return;
                setRssList(topicResp?.rss ?? []);
                hasFetched.current = true;
            })
            .catch((err) => {
                if (!cancelled) {
                    // Go errors land directly in catch
                    setError(typeof err === 'string' ? err : err?.message || 'Failed to fetch topic');
                }
            })
            .finally(() => {
                if (!cancelled) setLoading(false);
            });

        return () => {
            cancelled = true;
        };
    }, [isDropDownOpen, topicId]);

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
                <ul className="dropdown-menu-list">
                    {loading && <li className="dropdown-item">Loading...</li>}
                    
                    {!loading && error && (
                        <li className="dropdown-item dropdown-error">{error}</li>
                    )}
                    
                    {!loading && !error && rssList.length === 0 && (
                        <li className="dropdown-item">No subscriptions yet</li>
                    )}
                    
                    {!loading && !error && rssList.map((rss) => (
                        <li key={rss.id} className="dropdown-item">
                            {rss.title}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}

export default TopicMenu;

