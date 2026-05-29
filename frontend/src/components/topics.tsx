import {useState, useEffect} from 'react';
import  './topic_menu.css'
interface TopicMenuProps{
    topic: string;
}
const subscribed =  ["rss1", "rss2"]

function TopicMenu({topic}: TopicMenuProps) {
    const [isDropDownOpen, setIsDropDownOpen] = useState(false);
    
    return(
        <>
            <div className="dropdown">
            <button 
                className="dropdown-toggle" 
                onClick={() => setIsDropDownOpen(!isDropDownOpen)}
            >
                <span className="topic-text">{topic}</span> 
                <span className={`arrow ${isDropDownOpen ? 'up' : 'down'}`}>▼</span> 
            </button>

            {isDropDownOpen && (
                <ul className="dropdown-menu-list">
                    {subscribed.map((rss, index) => (
                        <li key={index} className="dropdown-item">
                            {rss}
                        </li>
                    ))}
                </ul>
            )}
            </div>
        </>
    );
}

export default TopicMenu
