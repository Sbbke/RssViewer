import {useState, useEffect} from 'react';
import {GetTopics} from '../../wailsjs/go/main/App';
import TopicMenu from './topics'
import './sidebar.css'

function Sidebar() {

    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [topics, setTopics] = useState<string[]>([]);

    useEffect(()=>{
        GetTopics().then((data:string[]) => {
                        setTopics(data);
        }).catch((err)=>{
            console.error("error loading topics", err)
        })
    }, [])

    return (
        <>
            <button 
                    className="mobile-toggle" 
                    onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            >
                        ☰
            </button>

            <aside className={`sidebar ${isSidebarOpen ? 'open' : 'closed'}`}>
                <div className="sidebar-brand"> Myapp </div>
                <div className="sidebar-menu">
                    <h3> main nevigation </h3>
                </div>
                 <ul className="topic-menu">
                        {topics.map((topic, index) => (
                            <li key={index} className="topic-item">
                               <TopicMenu topic={topic}/>   
                            </li>
                        ))}           
                </ul>                             
             </aside>
         </>       
    )
}
export default Sidebar
