import './rss_row.css';
interface RssRowProps {
    title: string;
    onRemove: () => void;
    removing: boolean;
    removeLabel: string;
    onSelect?: () => void;
}
 
function RssRow({ title, onRemove, removing, removeLabel, onSelect }: RssRowProps) {
    return (
        <li className="rss-row">
            {onSelect ? (
                <button className="rss-row-title" onClick={onSelect}>
                    {title}
                </button>
            ) : (
                <span className="rss-row-title">{title}</span>
            )}
            <button
                className="rss-unlink-btn"
                onClick={onRemove}
                disabled={removing}
                aria-label={removeLabel}
            >
                {removing ? '...' : '✕'}
            </button>
        </li>
    );
}
 
export default RssRow;
 
