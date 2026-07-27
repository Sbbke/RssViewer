import { useCallback, useState } from 'react';
 
/**
 * Wraps an async function (usually a Wails-generated API call) with
 * standard loading + error state, so components stop re-writing the
 * same .then().catch().finally() block over and over.
 *
 * Usage:
 *   const createTopic = useAsyncAction(CreateTopic);
 *   const created = await createTopic.run(name); // undefined on failure
 *   if (created) { ... }
 */
export function useAsyncAction<Args extends unknown[], R>(
    action: (...args: Args) => Promise<R>,
    fallbackMessage = 'Something went wrong. Please try again.'
) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
 
    const run = useCallback(
        async (...args: Args): Promise<R | undefined> => {
            setLoading(true);
            setError('');
            try {
                return await action(...args);
            } catch (err) {
                const message =
                    typeof err === 'string'
                        ? err
                        : (err as { message?: string })?.message || fallbackMessage;
                setError(message);
                return undefined;
            } finally {
                setLoading(false);
            }
        },
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [action]
    );
 
    return { run, loading, error, setError };
}
