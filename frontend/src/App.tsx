import { useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import Sidebar from './components/sidebar';
import TopicPageContainer from './components/content/topic_page/TopicPageContainer';
import GoldParticles from './components/background/GoldParticles';

// Legacy: components/content/topic_content.tsx (the flat feed-list
// view) is kept on disk but no longer wired up here — superseded by
// TopicPageContainer / TopicPage (header + briefing carousel +
// posts). Re-import it here if it's needed again.

interface SelectedTopic {
    id: number;
    name: string;
}

function App() {
    const [selectedTopic, setSelectedTopic] = useState<SelectedTopic | null>(null);
    return (
        <div id="app">
            <GoldParticles />
            <div className="app-container">
                <Sidebar
                    selectedTopicId={selectedTopic?.id ?? null}
                    onSelectTopic={(id, topicName) => setSelectedTopic({ id, name: topicName })}
                    onClearSelection={() => setSelectedTopic(null)}
                />
                <main id="App" className="main-content">
                    {selectedTopic ? (
                        <TopicPageContainer
                            topicId={selectedTopic.id}
                            topicName={selectedTopic.name}
                            onBack={() => setSelectedTopic(null)}
                            onOpenPost={(postId, rssId) => {
                                // TODO: wire to a post-detail view once one exists —
                                // for now this is a no-op placeholder, same as the
                                // legacy topic_content.tsx's onSelectPost.
                                console.log('open post', postId, 'from rss', rssId);
                            }}
                        />
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
        </div>
    );
}

export default App;
