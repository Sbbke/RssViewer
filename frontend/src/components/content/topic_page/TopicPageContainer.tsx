import { useCallback, useEffect, useState } from 'react';
import { GetTopicDetail } from '../../../../wailsjs/go/main/App';
import type { dto } from '../../../../wailsjs/go/models';
import ErrorModal from '../../common/ErrorModal';
import TopicPage from './TopicPage';

interface TopicPageContainerProps {
    topicId: number;
    topicName: string;
    topicDescription?: string;
    onBack: () => void;
    onOpenPost: (postId: number, rssId: number) => void;
}

function TopicPageContainer({
    topicId,
    topicName,
    topicDescription,
    onBack,
    onOpenPost,
}: TopicPageContainerProps) {
    const [data, setData] = useState<dto.TopicAllInOne | null>(null);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [error, setError] = useState('');

    const fetchDetail = useCallback((showFullPageSpinner: boolean) => {
        let cancelled = false;
        if (showFullPageSpinner) setLoading(true);
        setError('');
        GetTopicDetail(topicId)
            .then((result: dto.TopicAllInOne) => {
                if (!cancelled) setData(result);
            })
            .catch((err) => {
                if (!cancelled) {
                    setError(typeof err === 'string' ? err : err?.message || 'Failed to load topic');
                }
            })
            .finally(() => {
                if (!cancelled) {
                    if (showFullPageSpinner) setLoading(false);
                    else setRefreshing(false);
                }
            });
        return () => { cancelled = true; };
    }, [topicId]);

    useEffect(() => {
        const cancel = fetchDetail(true);
        return cancel;
    }, [fetchDetail]);

    const handleRefresh = () => {
        setRefreshing(true);
        fetchDetail(false);
    };

    if (loading) return <div className="content-loading">Loading {topicName}…</div>;

    if (error) {
        return <ErrorModal title="Couldn't load topic" message={error} onClose={onBack} />;
    }

    if (!data) return null;

    return (
        <TopicPage
            topicName={topicName}
            topicDescription={topicDescription}
            data={data}
            onBack={onBack}
            onOpenPost={onOpenPost}
            onRefresh={handleRefresh}
            isRefreshing={refreshing}
        />
    );
}

export default TopicPageContainer;
