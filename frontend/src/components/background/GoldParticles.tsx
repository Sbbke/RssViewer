import { useEffect, useRef } from 'react';
import './gold_particles.css';

interface Particle {
    x: number;
    y: number;
    size: number;
    speedY: number;
    speedX: number;
    rotation: number;
    rotSpeed: number;
    flipPhase: number;
    flipSpeed: number;
    opacity: number;
    color: string;
}

const PARTICLE_COUNT = 80;

function createParticle(canvas: HTMLCanvasElement): Particle {
    const size = Math.random() * 6 + 3;
    return {
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        size,
        speedY: -(Math.random() * 0.25 + 0.08),
        speedX: (Math.random() - 0.5) * 0.18,
        rotation: Math.random() * Math.PI * 2,
        rotSpeed: (Math.random() - 0.5) * 0.012,
        // continuous flip: phase advances every frame,
        // scaleY = cos(phase) sweeps smoothly 1 -> 0 -> -1 -> 0 -> 1 ...
        flipPhase: Math.random() * Math.PI * 2,
        flipSpeed: (Math.random() * 0.02 + 0.01) * (Math.random() > 0.5 ? 1 : -1),
        opacity: Math.random() * 0.55 + 0.25,
        color: Math.random() > 0.3 ? '#ffcd84' : '#f4c368',
    };
}

function drawTriangle(ctx: CanvasRenderingContext2D, p: Particle) {
    ctx.save();
    ctx.translate(p.x, p.y);
    ctx.rotate(p.rotation);

    // continuous pseudo-3D flip along Y axis
    const flipScale = Math.cos(p.flipPhase);
    ctx.scale(1, flipScale);

    // slight brightness/opacity dip as it turns edge-on, sells the 3D illusion
    ctx.globalAlpha = p.opacity * (0.4 + 0.6 * Math.abs(flipScale));

    ctx.fillStyle = p.color;
    ctx.beginPath();
    ctx.moveTo(0, -p.size);
    ctx.lineTo(p.size * 0.9, p.size * 0.7);
    ctx.lineTo(-p.size * 0.9, p.size * 0.7);
    ctx.closePath();
    ctx.fill();
    ctx.restore();
}

/**
 * Fixed full-viewport canvas of slowly drifting, continuously
 * flipping gold triangles. Mount once near the root (e.g. in
 * App.tsx) so it sits behind everything else.
 */
function GoldParticles() {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        let particles: Particle[] = [];
        let rafId = 0;

        function resize() {
            if (!canvas) return;
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
        }
        window.addEventListener('resize', resize);
        resize();

        particles = Array.from({ length: PARTICLE_COUNT }, () => createParticle(canvas));

        function animate() {
            if (!canvas || !ctx) return;
            ctx.clearRect(0, 0, canvas.width, canvas.height);

            particles.forEach((p) => {
                p.x += p.speedX;
                p.y += p.speedY;
                p.rotation += p.rotSpeed;
                p.flipPhase += p.flipSpeed;

                if (p.y < -20) {
                    p.y = canvas.height + 20;
                    p.x = Math.random() * canvas.width;
                }
                if (p.x < -20) p.x = canvas.width + 20;
                if (p.x > canvas.width + 20) p.x = -20;

                drawTriangle(ctx, p);
            });

            rafId = requestAnimationFrame(animate);
        }
        animate();

        return () => {
            window.removeEventListener('resize', resize);
            cancelAnimationFrame(rafId);
        };
    }, []);

    return <canvas ref={canvasRef} id="gold-particles" className="gold-particles-canvas" />;
}

export default GoldParticles;
