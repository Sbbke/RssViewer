import './topic_menu.css';
 
interface TopicMenuProps {
    topicId: number;
    topic: string; // display name
    isActive: boolean;
    onSelect: () => void;
}
 
/**
 * Previously this owned its own dropdown + feed list + add/remove
 * logic. That's all moved to <TopicContent>, which renders in the
 * main pane once a topic is selected — this component now only
 * needs to say "the user picked this topic."
 */
function TopicMenu({ topic, isActive, onSelect }: TopicMenuProps) {
    return (
        <button
            type="button"
            className={`topic-select-btn${isActive ? ' active' : ''}`}
            onClick={onSelect}
            aria-current={isActive ? 'true' : undefined}
        >
            <span className="topic-text">{topic}</span>
        </button>
    );
}
 
export default TopicMenu;
