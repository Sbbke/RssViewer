import type { ReactNode } from 'react';
import './confirm_modal.css';
 
interface ConfirmModalProps {
    title: string;
    message: ReactNode;
    confirmLabel?: string;
    cancelLabel?: string;
    danger?: boolean;
    busy?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
}
 
/**
 * One modal used for every "are you sure?" moment in the app —
 * deleting a topic, deleting a feed, unlinking a feed. Previously
 * this JSX was duplicated twice in sidebar.tsx and a third variant
 * (window.confirm) lived in topic_menu.tsx.
 */
function ConfirmModal({
    title,
    message,
    confirmLabel = 'Confirm',
    cancelLabel = 'Cancel',
    danger = true,
    busy = false,
    onConfirm,
    onCancel,
}: ConfirmModalProps) {
    return (
        <div className="modal-overlay" onClick={onCancel}>
            <div
                className="modal-content"
                role="dialog"
                aria-modal="true"
                aria-labelledby="confirm-modal-title"
                onClick={(e) => e.stopPropagation()}
            >
                <h3 id="confirm-modal-title">{title}</h3>
                <p>{message}</p>
                <div className="modal-actions">
                    <button className="btn-cancel" onClick={onCancel} disabled={busy}>
                        {cancelLabel}
                    </button>
                    <button
                        className={danger ? 'btn-danger' : 'btn-primary'}
                        onClick={onConfirm}
                        disabled={busy}
                        autoFocus
                    >
                        {busy ? 'Working…' : confirmLabel}
                    </button>
                </div>
            </div>
        </div>
    );
}
 
export default ConfirmModal;
