import { useState, useEffect } from 'react';
import { GetTopics, CreateTopic, DeleteTopic, SubmitRssUrl } from '../../wailsjs/go/main/App';
import type { dto } from '../../wailsjs/go/models';
import TopicMenu from './topic_menu';
import './sidebar.css';

function Sidebar() {
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [topics, setTopics] = useState<dto.TopicResponse[]>([]);
    const [error, setError] = useState<string>('');

    const [isAdding, setIsAdding] = useState(false);
    const [newTopicName, setNewTopicName] = useState('');
    const [creating, setCreating] = useState(false);

const [topicToDelete, setTopicToDelete] = useState<{ id: number; name: string } | null>(null);
    const [isAddingStandaloneRss, setIsAddingStandaloneRss] = useState(false);
    const [standaloneRssUrl, setStandaloneRssUrl] = useState('');
    const [submittingStandalone, setSubmittingStandalone] = useState(false);
const confirmDeleteTopic = () => {
    if (!topicToDelete) return;

    DeleteTopic(topicToDelete.id)
        .then(() => {
            setTopics((prev) => prev.filter((t) => t.topicId !== topicToDelete.id));
            setError('');
            setTopicToDelete(null); // Close modal
        })
        .catch((err) => {
            console.error("Error deleting topic:", err);
            setError(typeof err === 'string' ? err : err?.message || 'Failed to delete topic');
            setTopicToDelete(null);
        });
};
    const handleAddStandaloneRss = (e: React.FormEvent) => {
        e.preventDefault();
        const url = standaloneRssUrl.trim();
        if (!url || submittingStandalone) return;

        setSubmittingStandalone(true);
        setError('');
        // null topicId => standalone feed, not linked to any topic
        SubmitRssUrl(url, null)
            .then(() => {
                setStandaloneRssUrl('');
                setIsAddingStandaloneRss(false);
                // Note: nothing in the sidebar currently lists standalone feeds —
                // you'll want a "GetStandaloneRss" query + a section here to show them.
            })
            .catch((err) => {
                setError(typeof err === 'string' ? err : err?.message || 'Failed to add feed');
            })
            .finally(() => setSubmittingStandalone(false));
    };
    useEffect(() => {
        loadTopics();
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

    const handleCreateTopic = (e: React.FormEvent) => {
        e.preventDefault();
        const name = newTopicName.trim();
        if (!name || creating) return;

        setCreating(true);
        CreateTopic(name)
            .then((created: dto.TopicResponse) => {
                setTopics((prev) => [...prev, created]);
                setNewTopicName('');
                setIsAdding(false);
                setError('');
            })
            .catch((err) => {
                console.error("Error creating topic:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to create topic');
            })
            .finally(() => setCreating(false));
    };

    const handleDeleteTopic = (topicId: number) => {
        DeleteTopic(topicId)
            .then(() => {
                setTopics((prev) => prev.filter((t) => t.topicId !== topicId));
                setError('');
            })
            .catch((err) => {
                console.error("Error deleting topic:", err);
                setError(typeof err === 'string' ? err : err?.message || 'Failed to delete topic');
            });
    };

    const cancelAdd = () => {
        setIsAdding(false);
        setNewTopicName('');
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

                <ul className="topic-menu">
                    {topics.length === 0 && !isAdding && (
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

            <div className="standalone-rss-section">
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
                <div className="topic-add-section">
                    {isAdding ? (
                        <form className="topic-add-form" onSubmit={handleCreateTopic}>
                            <input
                                type="text"
                                autoFocus
                                value={newTopicName}
                                onChange={(e) => setNewTopicName(e.target.value)}
                                placeholder="Topic name"
                                disabled={creating}
                                onKeyDown={(e) => {
                                    if (e.key === 'Escape') cancelAdd();
                                }}
                            />
                            <div className="topic-add-actions">
                                <button
                                    type="submit"
                                    disabled={creating || !newTopicName.trim()}
                                >
                                    {creating ? 'Adding...' : 'Add'}
                                </button>
                                <button
                                    type="button"
                                    onClick={cancelAdd}
                                    disabled={creating}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    ) : (
                        <button
                            className="topic-add-btn"
                            onClick={() => setIsAdding(true)}
                        >
                            + Add Topic
                        </button>
                    )}
                </div>
            </aside>
        </>
    );}

export default Sidebar;
