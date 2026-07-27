interface RssRowProps {
    title: string;
    onRemove: () => void;
    removing: boolean;
    removeLabel: string;
}
 
function RssRow({ title, onRemove, removing, removeLabel }: RssRowProps) {
    return (
        <li className="dropdown-item rss-item">
            <span className="rss-title" title={title}>
                {title}
            </span>
            <button
                className="rss-unlink-btn"
                onClick={onRemove}
                disabled={removing}
                aria-label={removeLabel}
            >
                {removing ? (
                    <span className="spinner-tiny" aria-hidden="true" />
                ) : (
                    '✕'
                )}
            </button>
        </li>
    );
}
 
export default RssRow;
 
