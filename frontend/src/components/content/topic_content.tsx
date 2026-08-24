import { useEffect, useState } from 'react';
import { GetTopic, SubmitRssUrl, UnlinkRssFromTopic } from '../../../wailsjs/go/main/App';
import type { dto } from '../../../wailsjs/go/models';
import { useAsyncAction } from '../hook/UseAsyncAction';
import ErrorBanner from '../common/ErrorBanner';
import ConfirmModal from '../common/ConfirmModal';
import AddRssModal from '../common/AddRssModal';
import RssRow from '../common/RssRow';
import RssContent from './rss_content';
import './topic_content.css';

interface TopicContentProps {
    topicId: number;
    topicName: string;
}

function TopicContent({ topicId, topicName }: TopicContentProps) {
    const [rssList, setRssList] = useState<dto.RssItem[]>([]);
    const [loaded, setLoaded] = useState(false);
    const [isAddingRss, setIsAddingRss] = useState(false);
    const [unlinkTarget, setUnlinkTarget] = useState<dto.RssItem | null>(null);

    // NEW — which feed's posts are being viewed, if any
    const [selectedRss, setSelectedRss] = useState<{ id: number; title: string } | null>(null);

    const getTopic = useAsyncAction(GetTopic, 'Failed to load this topic.');
    const submitRss = useAsyncAction(SubmitRssUrl, 'Failed to add feed.');
    const unlinkRss = useAsyncAction(UnlinkRssFromTopic, 'Failed to remove feed.');

    const error = getTopic.error || submitRss.error || unlinkRss.error;
    const clearErrors = () => {
        getTopic.setError('');
        submitRss.setError('');
        unlinkRss.setError('');
    };

    useEffect(() => {
        setLoaded(false);
        setSelectedRss(null); // switching topics resets any drilled-in feed view
        getTopic.run(topicId).then((resp) => {
            setRssList(resp?.rss ?? []);
            setLoaded(true);
        });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [topicId]);

    const handleAddRss = async (url: string): Promise<boolean> => {
        const item = await submitRss.run(url, topicId);
        if (!item) return false;
        setRssList((prev) => [item, ...prev]);
        return true;
    };

    const handleConfirmUnlink = async () => {
        if (!unlinkTarget) return;
        const ok = await unlinkRss.run(unlinkTarget.id, topicId);
        if (ok !== undefined) {
            setRssList((prev) => prev.filter((r) => r.id !== unlinkTarget.id));
            // If the feed being viewed was just unlinked, back out to the list.
            if (selectedRss?.id === unlinkTarget.id) {
                setSelectedRss(null);
            }
        }
        setUnlinkTarget(null);
    };

    // Drilled into a specific feed's posts.
    if (selectedRss) {
        return (
            <div className="topic-content">
                <div className="topic-content-header">
                    <button className="back-btn" onClick={() => setSelectedRss(null)}>
                        ← Back to {topicName}
                    </button>
                </div>
                <RssContent
                    rssId={selectedRss.id}
                    rssTitle={selectedRss.title}
                    onSelectPost={(postId, postTitle) => {
                        console.log('selected post', postId, postTitle);
                    }}
                    onBack={() => setSelectedRss(null)}
                />
            </div>
        );
    }

    return (
        <div className="topic-content">
            <div className="topic-content-header">
                <h2>{topicName}</h2>
                <button className="rss-add-btn" onClick={() => setIsAddingRss(true)}>
                    + Add RSS
                </button>
            </div>

            <ErrorBanner message={error} onDismiss={clearErrors} />

            {!loaded && (
                <div className="topic-content-loading">
                    <span className="spinner-tiny" aria-hidden="true" /> Loading feeds…
                </div>
            )}

            {loaded && rssList.length === 0 && (
                <div className="topic-content-empty">
                    No feeds in <strong>{topicName}</strong> yet. Add one to get started.
                </div>
            )}

            {loaded && rssList.length > 0 && (
                <ul className="topic-content-feed-list">
                    {rssList.map((rss) => (
                        <RssRow
                            key={rss.id}
                            title={rss.title}
                            removing={unlinkRss.loading && unlinkTarget?.id === rss.id}
                            removeLabel={`Remove ${rss.title} from ${topicName}`}
                            onRemove={() => setUnlinkTarget(rss)}
                            onSelect={() => setSelectedRss({ id: rss.id, title: rss.title })}
                        />
                    ))}
                </ul>
            )}

            {isAddingRss && (
                <AddRssModal onSubmit={handleAddRss} onClose={() => setIsAddingRss(false)} />
            )}

            {unlinkTarget && (
                <ConfirmModal
                    title="Remove feed from topic"
                    message={
                        <>
                            Remove <strong>{unlinkTarget.title}</strong> from{' '}
                            <strong>{topicName}</strong>? The feed itself won&apos;t be
                            deleted — only its link to this topic.
                        </>
                    }
                    confirmLabel="Remove"
                    busy={unlinkRss.loading}
                    onConfirm={handleConfirmUnlink}
                    onCancel={() => setUnlinkTarget(null)}
                />
            )}
        </div>
    );
}

export default TopicContent;
