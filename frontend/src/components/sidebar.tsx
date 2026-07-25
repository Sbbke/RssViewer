import React, { useState, useEffect } from 'react';
import type { dto } from '../../wailsjs/go/models';
import TopicMenu from './topic_menu';
import './sidebar.css';
import { 
    GetTopics, 
    CreateTopic, 
    DeleteTopic, 
    SubmitRssUrl, 
    GetStandaloneRss, 
    RemoveRss 
} from '../../wailsjs/go/main/App';

function Sidebar() {
    // Sidebar visibility state
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [error, setError] = useState<string>('');

    // Topics state
    const [topics, setTopics] = useState<dto.TopicResponse[]>([]);
    const [isAddingTopic, setIsAddingTopic] = useState(false);
    const [newTopicName, setNewTopicName] = useState('');
    const [creatingTopic, setCreatingTopic] = useState(false);
    const [topicToDelete, setTopicToDelete] = useState<{ id: number; name: string } | null>(null);

    // Standalone RSS state
    const [standaloneRss, setStandaloneRss] = useState<dto.RssItem[]>([]);
    const [isAddingStandaloneRss, setIsAddingStandaloneRss] = useState(false);
    const [standaloneRssUrl, setStandaloneRssUrl] = useState('');
    const [submittingStandalone, setSubmittingStandalone] = useState(false);
    const [rssToDelete, setRssToDelete] = useState<{ id: number; title: string } | null>(null);

    // Fetch initial data on mount
    useEffect(() => {
        loadTopics();
        loadStandaloneRss();
    }, []);

    const loadTopics = () => {
        GetTopics()
            .then((data: dto.TopicResponse[]) => {
                setTopics(data ?? []);
                setError('');
            })
            .catch((err) => {
                console.error("Error loading topics:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to load topics');
            });
    };

    const loadStandaloneRss = () => {
        GetStandaloneRss()
            .then((data: dto.RssItem[]) => setStandaloneRss(data ?? []))
            .catch((err) => {
                console.error("Error loading standalone rss:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to load feeds');
            });
    };

    // Topic handlers
    const handleCreateTopic = (e: React.FormEvent) => {
        e.preventDefault();
        const name = newTopicName.trim();
        if (!name || creatingTopic) return;

        setCreatingTopic(true);
        CreateTopic(name)
            .then((created: dto.TopicResponse) => {
                setTopics((prev) => [...prev, created]);
                setNewTopicName('');
                setIsAddingTopic(false);
                setError('');
            })
            .catch((err) => {
                console.error("Error creating topic:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to create topic');
            })
            .finally(() => setCreatingTopic(false));
    };

    const confirmDeleteTopic = () => {
        if (!topicToDelete) return;

        DeleteTopic(topicToDelete.id)
            .then(() => {
                setTopics((prev) => prev.filter((t) => t.topicId !== topicToDelete.id));
                setError('');
                setTopicToDelete(null);
            })
            .catch((err) => {
                console.error("Error deleting topic:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to delete topic');
                setTopicToDelete(null);
            });
    };

    const cancelAddTopic = () => {
        setIsAddingTopic(false);
        setNewTopicName('');
    };

    // RSS handlers
    const handleAddStandaloneRss = (e: React.FormEvent) => {
        e.preventDefault();
        const url = standaloneRssUrl.trim();
        if (!url || submittingStandalone) return;

        setSubmittingStandalone(true);
        setError('');
        SubmitRssUrl(url, null)
            .then((item: dto.RssItem) => {
                setStandaloneRss((prev) => [item, ...prev]);
                setStandaloneRssUrl('');
                setIsAddingStandaloneRss(false);
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to add feed');
            })
            .finally(() => setSubmittingStandalone(false));
    };

    const confirmDeleteRss = () => {
        if (!rssToDelete) return;
        RemoveRss(rssToDelete.id)
            .then(() => {
                setStandaloneRss((prev) => prev.filter((r) => r.id !== rssToDelete.id));
                setError('');
                setRssToDelete(null);
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to delete feed');
                setRssToDelete(null);
            });
    };

    return (
        <>
            <button
                className="mobile-toggle"
                onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            >
                ☰
            </button>

            <aside className={`sidebar ${isSidebarOpen ? 'open' : 'closed'}`}>
                <div className="sidebar-brand">Myapp</div>
                <div className="sidebar-menu">
                    <h3>Main Navigation</h3>
                </div>

                {error && <div className="sidebar-error">{error}</div>}

                {/* Topics List */}
                <ul className="topic-menu">
                    {topics.length === 0 && !isAddingTopic && (
                        <li className="topic-empty">No topics yet</li>
                    )}
                    {topics.map((t) => (
                        <li key={t.topicId} className="topic-item">
                            <TopicMenu topicId={t.topicId} topic={t.name} />
                            <button
                                className="topic-delete-btn"
                                onClick={() => setTopicToDelete({ id: t.topicId, name: t.name })}
                                aria-label={`Delete ${t.name}`}
                            >
                                ✕
                            </button>
                        </li>
                    ))}
                </ul>

                {/* Topic Add Section */}
                <div className="topic-add-section">
                    {isAddingTopic ? (
                        <form className="topic-add-form" onSubmit={handleCreateTopic}>
                            <input
                                type="text"
                                autoFocus
                                value={newTopicName}
                                onChange={(e) => setNewTopicName(e.target.value)}
                                placeholder="Topic name"
                                disabled={creatingTopic}
                                onKeyDown={(e) => {
                                    if (e.key === 'Escape') cancelAddTopic();
                                }}
                            />
                            <div className="topic-add-actions">
                                <button
                                    type="submit"
                                    disabled={creatingTopic || !newTopicName.trim()}
                                >
                                    {creatingTopic ? 'Adding...' : 'Add'}
                                </button>
                                <button
                                    type="button"
                                    onClick={cancelAddTopic}
                                    disabled={creatingTopic}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    ) : (
                        <button
                            className="topic-add-btn"
                            onClick={() => setIsAddingTopic(true)}
                        >
                            + Add Topic
                        </button>
                    )}
                </div>

                {/* Standalone RSS Section */}
                <div className="standalone-rss-section">
                    <h4>Standalone Feeds</h4>
                    <ul className="standalone-rss-list">
                        {standaloneRss.length === 0 && (
                            <li className="topic-empty">No standalone feeds</li>
                        )}
                        {standaloneRss.map((r) => (
                            <li key={r.id} className="rss-item">
                                <span>{r.title}</span>
                                <button
                                    className="rss-unlink-btn"
                                    onClick={() => setRssToDelete({ id: r.id, title: r.title })}
                                    aria-label={`Delete ${r.title}`}
                                >
                                    ✕
                                </button>
                            </li>
                        ))}
                    </ul>

                    {isAddingStandaloneRss ? (
                        <form className="rss-add-form" onSubmit={handleAddStandaloneRss}>
                            <input
                                type="url"
                                autoFocus
                                value={standaloneRssUrl}
                                onChange={(e) => setStandaloneRssUrl(e.target.value)}
                                placeholder="https://example.com/feed.xml"
                                disabled={submittingStandalone}
                                onKeyDown={(e) => {
                                    if (e.key === 'Escape') {
                                        setIsAddingStandaloneRss(false);
                                        setStandaloneRssUrl('');
                                    }
                                }}
                            />
                            <div className="rss-add-actions">
                                <button type="submit" disabled={submittingStandalone || !standaloneRssUrl.trim()}>
                                    {submittingStandalone ? 'Adding...' : 'Add'}
                                </button>
                                <button
                                    type="button"
                                    onClick={() => { setIsAddingStandaloneRss(false); setStandaloneRssUrl(''); }}
                                    disabled={submittingStandalone}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    ) : (
                        <button
                            className="rss-add-btn"
                            onClick={() => setIsAddingStandaloneRss(true)}
                        >
                            + Add standalone RSS
                        </button>
                    )}
                </div>
            </aside>

            {/* Topic Deletion Modal */}
            {topicToDelete && (
                <div className="modal-overlay">
                    <div className="modal-content">
                        <h3>Confirm Deletion</h3>
                        <p>Are you sure you want to delete <strong>{topicToDelete.name}</strong>?</p>
                        <div className="modal-actions">
                            <button 
                                className="btn-cancel" 
                                onClick={() => setTopicToDelete(null)}
                            >
                                Cancel
                            </button>
                            <button 
                                className="btn-danger" 
                                onClick={confirmDeleteTopic}
                            >
                                Delete
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* RSS Deletion Modal */}
            {rssToDelete && (
                <div className="modal-overlay">
                    <div className="modal-content">
                        <h3>Confirm Deletion</h3>
                        <p>Delete <strong>{rssToDelete.title}</strong> permanently? This removes it everywhere, including any topics it's linked to.</p>
                        <div className="modal-actions">
                            <button className="btn-cancel" onClick={() => setRssToDelete(null)}>Cancel</button>
                            <button className="btn-danger" onClick={confirmDeleteRss}>Delete</button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
}

export default Sidebar;
