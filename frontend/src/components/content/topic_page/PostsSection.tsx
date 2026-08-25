import { useMemo, useState } from 'react';
import PostCard from './PostCard';
import FilterDrawer, { type FilterState } from './FilterDrawer';
import "./posts_section.css";

export interface TopicPost {
    id: number;
    title: string;
    excerpt?: string;
    source: string;
    publishedAt: string; // ISO date string
    thumbnail?: string;
}

interface PostsSectionProps {
    posts: TopicPost[];
    onOpenPost: (postId: number) => void;
    layout?: 'list' | 'grid';
    onRefresh?: () => void;
    isRefreshing?: boolean;
}

const DAY_MS = 24 * 60 * 60 * 1000;
const MAX_DISPLAYED = 10;

function withinDateFilter(publishedAt: string, filter: string): boolean {
    if (filter === 'all') return true;
    const age = Date.now() - new Date(publishedAt).getTime();
    if (filter === '24h') return age <= DAY_MS;
    if (filter === '7d') return age <= 7 * DAY_MS;
    if (filter === '30d') return age <= 30 * DAY_MS;
    return true;
}

function formatDateLabel(iso: string): string {
    return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}

function applyFilter(posts: TopicPost[], filter: FilterState): TopicPost[] {
    return posts.filter((p) => {
        const sourceOk = filter.selectedSources.size === 0 || filter.selectedSources.has(p.source);
        const dateOk = withinDateFilter(p.publishedAt, filter.dateFilter);
        return sourceOk && dateOk;
    });
}

function mostRecentFirst(posts: TopicPost[]): TopicPost[] {
    return [...posts].sort(
        (a, b) => new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime()
    );
}

function PostsSection({ posts, onOpenPost, layout = 'list', onRefresh, isRefreshing = false }: PostsSectionProps) {
    const [drawerOpen, setDrawerOpen] = useState(false);
    // The only filter state that actually affects what's rendered —
    // changes only land here via FilterDrawer's onApply, never live
    // while the drawer is open.
    const [applied, setApplied] = useState<FilterState>({
        selectedSources: new Set(),
        dateFilter: 'all',
    });

    const sources = useMemo(
        () => Array.from(new Set(posts.map((p) => p.source))).sort(),
        [posts]
    );

    const filteredPosts = useMemo(() => applyFilter(posts, applied), [posts, applied]);

    // Cap to the 10 most recent (post-filter) — kept as a single
    // short page, no infinite scroll here.
    const displayedPosts = useMemo(
        () => mostRecentFirst(filteredPosts).slice(0, MAX_DISPLAYED),
        [filteredPosts]
    );

    const filtersActive = applied.selectedSources.size > 0 || applied.dateFilter !== 'all';
    const hasMoreThanShown = filteredPosts.length > MAX_DISPLAYED;

    return (
        <section className="posts-section">
            <div className="posts-section-header">
                <h2 className="posts-section-title">
                    Posts{' '}
                    <span className="posts-section-count">
                        {hasMoreThanShown
                            ? `${displayedPosts.length} most recent of ${filteredPosts.length}`
                            : `${displayedPosts.length} ${displayedPosts.length === 1 ? 'result' : 'results'}`}
                    </span>
                </h2>
                <div className="posts-section-actions">
                    {onRefresh && (
                        <button
                            className="refresh-btn"
                            onClick={onRefresh}
                            disabled={isRefreshing}
                            aria-label="Refresh posts"
                        >
                            {isRefreshing ? (
                                <>
                                    <span className="spinner-tiny" aria-hidden="true" /> Refreshing…
                                </>
                            ) : (
                                '⟳ Refresh'
                            )}
                        </button>
                    )}
                    <button
                        className={`posts-filter-btn ${filtersActive ? 'active' : ''}`}
                        onClick={() => setDrawerOpen(true)}
                    >
                        ⚙ Filter{filtersActive ? ' •' : ''}
                    </button>
                </div>
            </div>

            {displayedPosts.length === 0 && (
                <p className="posts-section-empty">No posts match the current filters.</p>
            )}

            <div className={`posts-list posts-list-${layout}`}>
                {displayedPosts.map((p) => (
                    <PostCard
                        key={p.id}
                        title={p.title}
                        excerpt={p.excerpt}
                        source={p.source}
                        dateLabel={formatDateLabel(p.publishedAt)}
                        thumbnail={p.thumbnail}
                        onOpen={() => onOpenPost(p.id)}
                    />
                ))}
            </div>

            <FilterDrawer
                open={drawerOpen}
                applied={applied}
                sources={sources}
                onApply={(next) => {
                    setApplied(next);
                    setDrawerOpen(false);
                }}
                onCancel={() => setDrawerOpen(false)}
                previewCount={(draftState) => applyFilter(posts, draftState).length}
            />
        </section>
    );
}

export default PostsSection;
