import { Injectable, signal, WritableSignal, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import Lenis from 'lenis';

/**
 * Animation service — initializes GSAP + Lenis and provides reusable animation helpers.
 * All animations are cleaned up via GSAP context on component destroy.
 */
@Injectable({ providedIn: 'root' })
export class AnimationService {
  private platformId = inject(PLATFORM_ID);
  private lenis: Lenis | null = null;

  readonly lenisReady: WritableSignal<boolean> = signal(false);
  readonly isPaused: WritableSignal<boolean> = signal(false);

  get isBrowser(): boolean {
    return isPlatformBrowser(this.platformId);
  }

  initLenis(): void {
    if (!this.isBrowser) return;

    gsap.registerPlugin(ScrollTrigger);

    this.lenis = new Lenis({
      lerp: 0.08,
      duration: 1.2,
    });

    gsap.ticker.add((time) => {
      this.lenis?.raf(time * 1000);
    });

    gsap.ticker.lagSmoothing(0);
    this.lenisReady.set(true);
  }

  pauseLenis(): void {
    this.lenis?.stop();
    this.isPaused.set(true);
  }

  resumeLenis(): void {
    this.lenis?.start();
    this.isPaused.set(false);
  }

  destroyLenis(): void {
    this.lenis?.destroy();
    this.lenis = null;
    this.lenisReady.set(false);
  }

  /** Create a GSAP context scoped to an element for easy cleanup */
  createContext(element: HTMLElement): gsap.Context {
    return gsap.context(() => {}, element);
  }

  /** Staggered fade-up entrance animation for a list of elements */
  fadeUpStagger(elements: string | Element | Element[], container?: Element): gsap.core.Tween {
    return gsap.from(elements, {
      y: 24,
      opacity: 0,
      duration: 0.6,
      stagger: 0.06,
      ease: 'power2.out',
      scrollTrigger: container ? {
        trigger: container,
        start: 'top 85%',
        toggleActions: 'play none none none',
      } : undefined,
    });
  }

  /** Count-up number tween */
  counterTween(target: { value: number }, endValue: number, duration = 1.4): gsap.core.Tween {
    return gsap.to(target, {
      value: endValue,
      duration,
      ease: 'power2.out',
      snap: { value: 1 },
    });
  }

  /** Repeating pulse animation for leader nodes */
  pulseElement(element: Element): gsap.core.Tween {
    return gsap.to(element, {
      scale: 1.06,
      duration: 1.2,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
    });
  }

  /** Sidebar icon hover effect */
  iconHover(element: Element, enter: boolean): gsap.core.Tween {
    return gsap.to(element, {
      x: enter ? 4 : 0,
      duration: 0.2,
      ease: 'power1.out',
    });
  }

  /** Button hover effect */
  buttonHover(element: Element, enter: boolean): gsap.core.Tween {
    return gsap.to(element, {
      y: enter ? -2 : 0,
      boxShadow: enter
        ? '0 8px 24px rgba(119, 126, 240, 0.4)'
        : '0 0 0 rgba(119, 126, 240, 0)',
      duration: 0.18,
      ease: 'power1.out',
    });
  }

  /** SVG line draw-on animation */
  lineDrawOn(paths: SVGPathElement[], duration = 1.5): gsap.core.Timeline {
    const tl = gsap.timeline();
    paths.forEach(path => {
      const length = path.getTotalLength();
      tl.fromTo(path,
        { strokeDasharray: length, strokeDashoffset: length },
        { strokeDashoffset: 0, duration, ease: 'power2.inOut' },
        0
      );
    });
    return tl;
  }

  /** Election animation timeline */
  electionTimeline(nodes: Element[], winnerId: number): gsap.core.Timeline {
    const tl = gsap.timeline();

    // All candidates flash yellow
    tl.to(nodes, {
      backgroundColor: '#FCD34D',
      duration: 0.3,
      stagger: 0.05,
    });

    // Randomise effect
    tl.to(nodes, {
      scale: 1.05,
      duration: 0.15,
      stagger: { each: 0.05, repeat: 3, yoyo: true },
    });

    // Reset all
    tl.to(nodes, {
      backgroundColor: 'var(--rw-core)',
      scale: 1,
      duration: 0.3,
    });

    // Crown winner
    tl.to(nodes[winnerId], {
      backgroundColor: 'var(--rw-accent)',
      scale: 1.1,
      boxShadow: '0 0 24px rgba(119, 126, 240, 0.6)',
      duration: 0.5,
      ease: 'back.out(1.7)',
    });

    return tl;
  }

  /** ScrollTrigger reveal for stat cards */
  scrollReveal(elements: string | Element | Element[], container: Element): ScrollTrigger[] {
    const triggers: ScrollTrigger[] = [];
    gsap.utils.toArray(elements).forEach((el: any) => {
      const trigger = ScrollTrigger.create({
        trigger: el,
        start: 'top 90%',
        onEnter: () => {
          gsap.from(el, {
            y: 24,
            opacity: 0,
            duration: 0.6,
            ease: 'power2.out',
          });
        },
        once: true,
      });
      triggers.push(trigger);
    });
    return triggers;
  }
}
