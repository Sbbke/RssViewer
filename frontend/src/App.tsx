import { useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import Sidebar from './components/sidebar';
import TopicContent from './components/content/topic_content';
 
interface SelectedTopic {
    id: number;
    name: string;
}
 
function App() {
    const [selectedTopic, setSelectedTopic] = useState<SelectedTopic | null>(null);

   return (
        <div className="app-container">
            <Sidebar
                selectedTopicId={selectedTopic?.id ?? null}
                onSelectTopic={(id, topicName) => setSelectedTopic({ id, name: topicName })}
                onClearSelection={() => setSelectedTopic(null)}
            />
            <main id="App" className="main-content">
                {selectedTopic ? (
                    <TopicContent topicId={selectedTopic.id} topicName={selectedTopic.name} />
                ) : (
                    <div className="main-placeholder">
                        <img src={logo} id="logo" alt="logo" />
 
                       <p className="main-placeholder-hint">
                            Select a topic from the sidebar to view its feeds.
                        </p>
                    </div>
                )}
            </main>
        </div>
    );
}
 
export default App;
