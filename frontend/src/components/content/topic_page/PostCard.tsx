import "./post_card.css";

interface PostCardProps {
    title: string;
    excerpt?: string;
    source: string;
    dateLabel: string;
    thumbnail?: string; // image src, caller decides encoding
    onOpen: () => void;
}

function PostCard({ title, excerpt, source, dateLabel, thumbnail, onOpen }: PostCardProps) {
    return (
        <button className="post-card" onClick={onOpen}>
            {thumbnail && (
                <img src={thumbnail} alt="" className="post-card-thumb" aria-hidden="true" />
            )}
            <div className="post-card-body">
                <h3 className="post-card-title">{title}</h3>
                {excerpt && <p className="post-card-excerpt">{excerpt}</p>}
                <div className="post-card-footer">
                    <span className="post-card-source">
                        {source} · {dateLabel}
                    </span>
                    <span className="post-card-arrow" aria-hidden="true">→</span>
                </div>
            </div>
        </button>
    );
}

export default PostCard;
