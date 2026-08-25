import { useEffect, useState } from 'react';
import "./filter_drawer.css";

export interface FilterState {
    selectedSources: Set<string>; // empty set = "all sources"
    dateFilter: string;
}

interface FilterDrawerProps {
    open: boolean;
    applied: FilterState;
    sources: string[];
    // Called only when the user clicks Apply — never on individual
    // checkbox/select changes.
    onApply: (next: FilterState) => void;
    // Called on Cancel/backdrop-click/Esc — discards any in-progress
    // draft changes, applied filters stay untouched.
    onCancel: () => void;
    // Given a candidate filter state, returns how many posts it would
    // match — used to show a live preview count inside the drawer
    // without touching the actual displayed list.
    previewCount: (state: FilterState) => number;
}

const DATE_OPTIONS = [
    { value: 'all', label: 'All time' },
    { value: '24h', label: 'Last 24 hours' },
    { value: '7d', label: 'Last 7 days' },
    { value: '30d', label: 'Last 30 days' },
];

function FilterDrawer({ open, applied, sources, onApply, onCancel, previewCount }: FilterDrawerProps) {
    const [draft, setDraft] = useState<FilterState>(applied);

    // Re-seed the draft from the currently-applied filters every time
    // the drawer opens, so stale edits from a previous open/cancel
    // don't linger.
    useEffect(() => {
        if (open) setDraft(applied);
    }, [open, applied]);

    if (!open) return null;

    const toggleSource = (source: string) => {
        setDraft((prev) => {
            const next = new Set(prev.selectedSources);
            if (next.has(source)) {
                next.delete(source);
            } else if (next.size === 0) {
                // Starting from "all selected" (empty set): clicking one
                // source means "only this one" — select everything else.
                sources.forEach((s) => { if (s !== source) next.add(s); });
            } else {
                next.add(source);
            }
            return { ...prev, selectedSources: next };
        });
    };

    const setDateFilter = (value: string) => {
        setDraft((prev) => ({ ...prev, dateFilter: value }));
    };

    const clearDraft = () => {
        setDraft({ selectedSources: new Set(), dateFilter: 'all' });
    };

    return (
        <div className="filter-drawer-overlay" onClick={onCancel}>
            <div
                className="filter-drawer"
                role="dialog"
                aria-label="Filter posts"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="filter-drawer-header">
                    <span className="filter-drawer-title">Posts</span>
                    <span className="filter-drawer-count">{previewCount(draft)} results</span>
                </div>

                {sources.length > 0 && (
                    <div className="filter-drawer-group">
                        <p className="filter-drawer-label">Source</p>
                        {sources.map((source) => (
                            <label key={source} className="filter-drawer-checkbox">
                                <input
                                    type="checkbox"
                                    checked={draft.selectedSources.size === 0 || draft.selectedSources.has(source)}
                                    onChange={() => toggleSource(source)}
                                />
                                {source}
                            </label>
                        ))}
                    </div>
                )}

                <div className="filter-drawer-group">
                    <p className="filter-drawer-label">Date</p>
                    <select
                        className="filter-drawer-select"
                        value={draft.dateFilter}
                        onChange={(e) => setDateFilter(e.target.value)}
                    >
                        {DATE_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>
                                {opt.label}
                            </option>
                        ))}
                    </select>
                </div>

                <div className="filter-drawer-actions">
                    <button className="btn-cancel" onClick={clearDraft}>
                        Clear
                    </button>
                    <button className="btn-primary" onClick={() => onApply(draft)}>
                        Apply
                    </button>
                </div>
            </div>
        </div>
    );
}

export default FilterDrawer;
