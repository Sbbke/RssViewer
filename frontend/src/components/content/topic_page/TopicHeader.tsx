import "./topic_header.css";

interface TopicHeaderProps {
    title: string;
    description?: string;
    postCount: number;
    lastUpdatedLabel?: string; // e.g. "2h ago" — caller formats this
}

function TopicHeader({ title, description, postCount, lastUpdatedLabel }: TopicHeaderProps) {
    return (
        <header className="topic-page-header">
            <h1 className="topic-page-title">{title}</h1>
            {description && <p className="topic-page-description">{description}</p>}
            <p className="topic-page-meta">
                {postCount} {postCount === 1 ? 'post' : 'posts'}
                {lastUpdatedLabel && <> · Last updated {lastUpdatedLabel}</>}
            </p>
        </header>
    );
}

export default TopicHeader;
