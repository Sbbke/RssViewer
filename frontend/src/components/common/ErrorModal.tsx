import type { ReactNode } from 'react';

interface ErrorModalProps {
    title?: string;
    message: ReactNode;
    onClose: () => void;
}

/**
 * A single-button popup for surfacing error messages. Reuses the same
 * .modal-overlay / .modal-content shell as ConfirmModal so it looks
 * consistent with the rest of the app's modals.
 *
 * Usage:
 *   const [error, setError] = useState('');
 *   ...
 *   {error && (
 *     <ErrorModal message={error} onClose={() => setError('')} />
 *   )}
 */
function ErrorModal({ title = 'Something went wrong', message, onClose }: ErrorModalProps) {
    return (
        <div className="modal-overlay" onClick={onClose}>
            <div
                className="modal-content"
                role="alertdialog"
                aria-modal="true"
                aria-labelledby="error-modal-title"
                onClick={(e) => e.stopPropagation()}
            >
                <h3 id="error-modal-title" className="modal-title">
                    {title}
                </h3>
                <p className="modal-message">{message}</p>
                <div className="modal-actions">
                    <button className="btn-primary" onClick={onClose} autoFocus>
                        OK
                    </button>
                </div>
            </div>
        </div>
    );
}

export default ErrorModal;
