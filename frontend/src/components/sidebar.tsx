import { useEffect, useState } from 'react';
import type { dto } from '../../wailsjs/go/models';
import TopicMenu from './topic_menu';
import './sidebar.css';
import {
    GetTopics,
    CreateTopic,
    DeleteTopic,
    SubmitRssUrl,
    GetStandaloneRss,
    RemoveRss,
} from '../../wailsjs/go/main/App';
import { useAsyncAction } from './hook/UseAsyncAction';
import ErrorBanner from './common/ErrorBanner';
import ConfirmModal from './common/ConfirmModal';
import InlineAddForm from './common/InlineAddForm';
import AddRssModal from './common/AddRssModal';
import RssRow from './common/RssRow';
import AddTopicModal from './common/AddTopicModal';

interface SidebarProps {
    selectedTopicId: number | null;
    onSelectTopic: (topicId: number, topicName: string) => void;
    onClearSelection: () => void; 
}

 
function Sidebar({ selectedTopicId, onSelectTopic,onClearSelection }: SidebarProps) {
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [isHovering, setIsHovering] = useState(false);
    const [topics, setTopics] = useState<dto.TopicResponse[]>([]);
    const [isAddingTopic, setIsAddingTopic] = useState(false);
    const [topicToDelete, setTopicToDelete] = useState<{ id: number; name: string } | null>(null);
 
    const [standaloneRss, setStandaloneRss] = useState<dto.RssItem[]>([]);
    const [isAddingStandaloneRss, setIsAddingStandaloneRss] = useState(false);
    const [rssToDelete, setRssToDelete] = useState<{ id: number; title: string } | null>(null);
 
    const [topicsLoaded, setTopicsLoaded] = useState(false);
    const [rssLoaded, setRssLoaded] = useState(false);
 
    const fetchTopics = useAsyncAction(GetTopics, 'Failed to load topics.');
    const fetchStandaloneRss = useAsyncAction(GetStandaloneRss, 'Failed to load feeds.');
    const createTopic = useAsyncAction(CreateTopic, 'Failed to create topic.');
    const deleteTopic = useAsyncAction(DeleteTopic, 'Failed to delete topic.');
    const submitStandaloneRss = useAsyncAction(SubmitRssUrl, 'Failed to add feed.');
    const removeRss = useAsyncAction(RemoveRss, 'Failed to delete feed.');
 
    const error =
        fetchTopics.error ||
        fetchStandaloneRss.error ||
        createTopic.error ||
        deleteTopic.error ||
        submitStandaloneRss.error ||
        removeRss.error;
 
    const clearErrors = () => {
        fetchTopics.setError('');
        fetchStandaloneRss.setError('');
        createTopic.setError('');
        deleteTopic.setError('');
        submitStandaloneRss.setError('');
        removeRss.setError('');
    };
 
    useEffect(() => {
        fetchTopics.run().then((data) => {
            setTopics(data ?? []);
            setTopicsLoaded(true);
        });
        fetchStandaloneRss.run().then((data) => {
            setStandaloneRss(data ?? []);
            setRssLoaded(true);
        });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);
 
    const handleCreateTopic = async (name: string): Promise<boolean> => {
        const created = await createTopic.run(name);
        if (!created) return false;
        setTopics((prev) => [...prev, created]);
        setIsAddingTopic(false);
        return true;
    };
 
    const handleConfirmDeleteTopic = async () => {
        if (!topicToDelete) return;
        const ok = await deleteTopic.run(topicToDelete.id);
        if (ok !== undefined) {
            setTopics((prev) => prev.filter((t) => t.topicId !== topicToDelete.id));
            // If the deleted topic was the one currently shown in the main pane,
            // reset the view so it doesn't keep pointing at a topic that no longer exists.
            if (selectedTopicId === topicToDelete.id) {
                onClearSelection();
            }
        }
        setTopicToDelete(null);
    };
 
    const handleAddStandaloneRss = async (url: string): Promise<boolean> => {
        const item = await submitStandaloneRss.run(url, null);
        if (!item) return false;
        setStandaloneRss((prev) => [item, ...prev]);
        return true;
    };
 
    const handleConfirmDeleteRss = async () => {
        if (!rssToDelete) return;
        const ok = await removeRss.run(rssToDelete.id);
        if (ok !== undefined) {
            setStandaloneRss((prev) => prev.filter((r) => r.id !== rssToDelete.id));
        }
        setRssToDelete(null);
    };
 
    return (
        <>
            <button
                className={`mobile-toggle ${isHovering ? 'is-hovering' : ''}`}
                onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                onMouseEnter={() => setIsHovering(true)}
                onMouseLeave={() => setIsHovering(false)}
                aria-label={isSidebarOpen ? 'Close sidebar' : 'Open sidebar'}
                aria-expanded={isSidebarOpen}
            >
                ☰
            </button>

            <aside className={`sidebar ${isSidebarOpen ? 'open' : 'closed'}`}>
                <ErrorBanner message={error} onDismiss={clearErrors} />
 
                {/* Topics — click a topic to view its feeds in the main pane */}
                <ul className="topic-menu">
                    {!topicsLoaded && (
                        <li className="topic-empty">
                            <span className="spinner-tiny" aria-hidden="true" /> Loading topics…
                        </li>
                    )}
                    {topicsLoaded && topics.length === 0 && !isAddingTopic && (
                        <li className="topic-empty">No topics yet — create one below.</li>
                    )}
                    {topics.map((t) => (
                        <li key={t.topicId} className="topic-item">
                            <TopicMenu
                                topicId={t.topicId}
                                topic={t.name}
                                isActive={selectedTopicId === t.topicId}
                                onSelect={() => onSelectTopic(t.topicId, t.name)}
                            />
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
 
                <div className="topic-add-section">
                    <button className="topic-add-btn" onClick={() => setIsAddingTopic(true)}>
                        + Add Topic
                    </button>
                </div>
                 
                {/* Standalone RSS */}
                <div className="standalone-rss-section">
                    {!rssLoaded ? (
                        <div className="standalone-rss-section">
                            <h4>Standalone Feeds</h4>
                            <ul className="standalone-rss-list">
                                <li className="topic-empty">
                                    <span className="spinner-tiny" aria-hidden="true" /> Loading feeds…
                                </li>
                            </ul>
                        </div>
                    ) : (
                        standaloneRss.length > 0 && (
                            <div className="standalone-rss-section">
                                <h4>Standalone Feeds</h4>
                                <ul className="standalone-rss-list">
                                    {standaloneRss.map((r) => (
                                        <RssRow
                                            key={r.id}
                                            title={r.title}
                                            removing={removeRss.loading && rssToDelete?.id === r.id}
                                            removeLabel={`Delete ${r.title}`}
                                            onRemove={() => setRssToDelete({ id: r.id, title: r.title })}
                                        />
                                    ))}
                                </ul>
                            </div>
                        )
                    )} 
                    <button
                        className="rss-add-btn"
                        onClick={() => setIsAddingStandaloneRss(true)}
                    >
                        + Add standalone RSS
                    </button>
                </div>
            </aside>
 
            {isAddingStandaloneRss && (
                <AddRssModal
                    onSubmit={handleAddStandaloneRss}
                    onClose={() => setIsAddingStandaloneRss(false)}
                />
            )}
             {isAddingTopic && (
                <AddTopicModal
                    onSubmit={handleCreateTopic}
                    onClose={() => setIsAddingTopic(false)}
                />
            )}
            {topicToDelete && (
                <ConfirmModal
                    title="Delete topic"
                    message={
                        <>
                            Are you sure you want to delete{' '}
                            <strong>{topicToDelete.name}</strong>?
                        </>
                    }
                    confirmLabel="Delete"
                    busy={deleteTopic.loading}
                    onConfirm={handleConfirmDeleteTopic}
                    onCancel={() => setTopicToDelete(null)}
                />
            )}
 
            {rssToDelete && (
                <ConfirmModal
                    title="Delete feed"
                    message={
                        <>
                            Delete <strong>{rssToDelete.title}</strong> permanently? This
                            removes it everywhere, including any topics it&apos;s linked to.
                        </>
                    }
                    confirmLabel="Delete"
                    busy={removeRss.loading}
                    onConfirm={handleConfirmDeleteRss}
                    onCancel={() => setRssToDelete(null)}
                />
            )}
        </>
    );
}
 
export default Sidebar;
