import { useCallback, useEffect, useState } from 'react';
import "./briefing_carousel.css";

interface BriefingCarouselProps {
    // Each entry is whatever the <img> src should be — caller decides
    // encoding (e.g. `data:image/png;base64,${bytes}` if the backend
    // sends raw slide bytes as base64 over JSON).
    slides: string[];
}

function BriefingCarousel({ slides }: BriefingCarouselProps) {
    const [index, setIndex] = useState(0);
    const total = slides.length;

    const goTo = useCallback((i: number) => {
        if (total === 0) return;
        setIndex(((i % total) + total) % total); // wrap both directions
    }, [total]);

    const goPrev = useCallback(() => goTo(index - 1), [goTo, index]);
    const goNext = useCallback(() => goTo(index + 1), [goTo, index]);

    useEffect(() => {
        const handleKey = (e: KeyboardEvent) => {
            if (e.key === 'ArrowLeft') goPrev();
            if (e.key === 'ArrowRight') goNext();
        };
        window.addEventListener('keydown', handleKey);
        return () => window.removeEventListener('keydown', handleKey);
    }, [goPrev, goNext]);

    if (total === 0) {
        return (
            <section className="briefing-carousel briefing-carousel-empty">
                <p>No briefing slides generated yet.</p>
            </section>
        );
    }

    const prevIndex = (index - 1 + total) % total;
    const nextIndex = (index + 1) % total;

    return (
        <section className="briefing-carousel" aria-roledescription="carousel" aria-label="Topic briefing slides">
            <div className="briefing-carousel-track">
                <button
                    className="briefing-carousel-arrow briefing-carousel-arrow-prev"
                    onClick={goPrev}
                    aria-label="Previous slide"
                    disabled={total <= 1}
                >
                    ‹
                </button>

                <div className="briefing-carousel-stage">
                    {total > 1 && (
                        <img
                            key={`prev-${prevIndex}`}
                            src={slides[prevIndex]}
                            alt=""
                            className="briefing-slide briefing-slide-peek briefing-slide-peek-left"
                            onClick={goPrev}
                            aria-hidden="true"
                        />
                    )}
                    <img
                        key={`current-${index}`}
                        src={slides[index]}
                        alt={`Briefing slide ${index + 1} of ${total}`}
                        className="briefing-slide briefing-slide-current"
                    />
                    {total > 1 && (
                        <img
                            key={`next-${nextIndex}`}
                            src={slides[nextIndex]}
                            alt=""
                            className="briefing-slide briefing-slide-peek briefing-slide-peek-right"
                            onClick={goNext}
                            aria-hidden="true"
                        />
                    )}
                </div>

                <button
                    className="briefing-carousel-arrow briefing-carousel-arrow-next"
                    onClick={goNext}
                    aria-label="Next slide"
                    disabled={total <= 1}
                >
                    ›
                </button>
            </div>

            <div className="briefing-carousel-footer">
                <div className="briefing-carousel-dots" role="tablist">
                    {slides.map((_, i) => (
                        <button
                            key={i}
                            className={`briefing-carousel-dot ${i === index ? 'active' : ''}`}
                            onClick={() => goTo(i)}
                            role="tab"
                            aria-selected={i === index}
                            aria-label={`Go to slide ${i + 1}`}
                        />
                    ))}
                </div>
                <span className="briefing-carousel-counter">
                    {index + 1} / {total}
                </span>
            </div>
        </section>
    );
}

export default BriefingCarousel;
