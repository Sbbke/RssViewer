import './error_banner.css';
 
interface ErrorBannerProps {
    message: string;
    onDismiss?: () => void;
}
 
/** Shows nothing when message is empty — safe to always render. */
function ErrorBanner({ message, onDismiss }: ErrorBannerProps) {
    if (!message) return null;
 
    return (
        <div className="error-banner" role="alert" aria-live="assertive">
            <span>{message}</span>
            {onDismiss && (
                <button
                    type="button"
                    className="error-banner-dismiss"
                    onClick={onDismiss}
                    aria-label="Dismiss error"
                >
                    ✕
                </button>
            )}
        </div>
    );
}
 
export default ErrorBanner;
