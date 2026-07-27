import { useState } from 'react';
import './add_rss_modal.css';
 
interface AddRssModalProps {
    title?: string;
    onSubmit: (url: string) => Promise<boolean>;
    onClose: () => void;
}
 
/**
 * A focused popup for the one thing it does: take a feed URL and
 * submit it. Closes itself automatically on success; stays open
 * (showing the caller's error, via ErrorBanner elsewhere) on failure
 * so the user can fix the URL and retry without retyping.
 */
function AddRssModal({ title = 'Add RSS feed', onSubmit, onClose }: AddRssModalProps) {
    const [url, setUrl] = useState('');
    const [submitting, setSubmitting] = useState(false);
 
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = url.trim();
        if (!trimmed || submitting) return;
 
        setSubmitting(true);
        const succeeded = await onSubmit(trimmed);
        setSubmitting(false);
        if (succeeded) onClose();
    };
 
    return (
        <div className="modal-overlay" onClick={submitting ? undefined : onClose}>
            <div
                className="modal-content"
                role="dialog"
                aria-modal="true"
                aria-labelledby="add-rss-modal-title"
                onClick={(e) => e.stopPropagation()}
            >
                <h3 id="add-rss-modal-title">{title}</h3>
                <form className="add-rss-form" onSubmit={handleSubmit}>
                    <input
                        type="url"
                        autoFocus
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        placeholder="https://example.com/feed.xml"
                        disabled={submitting}
                        onKeyDown={(e) => {
                            if (e.key === 'Escape') onClose();
                        }}
                    />
                    <div className="modal-actions">
                        <button
                            type="button"
                            className="btn-cancel"
                            onClick={onClose}
                            disabled={submitting}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="btn-primary"
                            disabled={submitting || !url.trim()}
                        >
                            {submitting ? 'Adding…' : 'Add feed'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
 
export default AddRssModal;
