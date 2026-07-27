import { useState } from 'react';
import './inline_add_form.css';
 
interface InlineAddFormProps {
    placeholder: string;
    inputType?: 'text' | 'url';
    submitLabel?: string;
    /** Return true on success (clears the field) or false to keep it. */
    onSubmit: (value: string) => Promise<boolean>;
    onCancel: () => void;
}
 
/**
 * The "type something, Add / Cancel" row used for new topics and new
 * RSS feeds. Handles its own text state + Escape-to-cancel; the
 * caller only supplies what "submit" actually does.
 */
function InlineAddForm({
    placeholder,
    inputType = 'text',
    submitLabel = 'Add',
    onSubmit,
    onCancel,
}: InlineAddFormProps) {
    const [value, setValue] = useState('');
    const [submitting, setSubmitting] = useState(false);
 
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = value.trim();
        if (!trimmed || submitting) return;
 
        setSubmitting(true);
        const succeeded = await onSubmit(trimmed);
        setSubmitting(false);
        if (succeeded) setValue('');
    };
 
    return (
        <form className="inline-add-form" onSubmit={handleSubmit}>
            <input
                type={inputType}
                autoFocus
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={placeholder}
                disabled={submitting}
                onKeyDown={(e) => {
                    if (e.key === 'Escape') onCancel();
                }}
            />
            <div className="inline-add-actions">
                <button type="submit" disabled={submitting || !value.trim()}>
                    {submitting ? 'Adding…' : submitLabel}
                </button>
                <button type="button" onClick={onCancel} disabled={submitting}>
                    Cancel
                </button>
            </div>
        </form>
    );
}
 
export default InlineAddForm;
 
