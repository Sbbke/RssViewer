import { useState } from 'react';

interface AddTopicModalProps {
    onSubmit: (name: string) => Promise<boolean>;
    onClose: () => void;
}

function AddTopicModal({ onSubmit, onClose }: AddTopicModalProps) {
    const [name, setName] = useState('');
    const [submitting, setSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = name.trim();
        if (!trimmed || submitting) return;

        setSubmitting(true);
        const ok = await onSubmit(trimmed);
        setSubmitting(false);
        if (ok) {
            setName('');
            // onClose is left to the caller — mirrors AddRssModal's contract
            // where a successful submit closes the modal (see handleCreateTopic).
        }
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <h3>Add topic</h3>
                <form onSubmit={handleSubmit}>
                    <input
                        type="text"
                        autoFocus
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="Topic name"
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
                            disabled={submitting || !name.trim()}
                        >
                            {submitting ? 'Adding...' : 'Add'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

export default AddTopicModal;
