import { useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import { Greet } from '../wailsjs/go/main/App';
import Sidebar from './components/sidebar';
import TopicContent from './components/content/topic_content';
 
interface SelectedTopic {
    id: number;
    name: string;
}
 
function App() {
    const [resultText, setResultText] = useState('Please enter your name below 👇');
    const [name, setName] = useState('');
    const [selectedTopic, setSelectedTopic] = useState<SelectedTopic | null>(null);
 
    const updateName = (e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);
 
    function greet() {
        Greet(name).then(updateResultText);
    }
 
    return (
        <div className="app-container">
            <Sidebar
                selectedTopicId={selectedTopic?.id ?? null}
                onSelectTopic={(id, topicName) => setSelectedTopic({ id, name: topicName })}
            />
            <main id="App" className="main-content">
                {selectedTopic ? (
                    <TopicContent topicId={selectedTopic.id} topicName={selectedTopic.name} />
                ) : (
                    <div className="main-placeholder">
                        <img src={logo} id="logo" alt="logo" />
                        <div id="result" className="result">
                            {resultText}
                        </div>
                        <div id="input" className="input-box">
                            <input
                                id="name"
                                className="input"
                                onChange={updateName}
                                autoComplete="off"
                                name="input"
                                type="text"
                            />
                            <button className="btn" onClick={greet}>
                                Greet
                            </button>
                        </div>
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
