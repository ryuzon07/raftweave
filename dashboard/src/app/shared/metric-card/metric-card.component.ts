import { Component, input, AfterViewInit, ElementRef, viewChild, OnDestroy } from '@angular/core';
import { gsap } from 'gsap';

@Component({
  selector: 'rw-metric-card',
  standalone: true,
  template: `
    <div class="metric-card" #card>
      <div class="metric-card__header">
        <span class="metric-card__label">{{ label() }}</span>
        @if (trend()) {
          <span class="metric-card__trend" [class.up]="trend() === 'up'" [class.down]="trend() === 'down'">
            {{ trend() === 'up' ? '↑' : trend() === 'down' ? '↓' : '→' }}
          </span>
        }
      </div>
      <div class="metric-card__value" #valueEl>{{ displayValue() }}</div>
      @if (subtitle()) {
        <div class="metric-card__subtitle">{{ subtitle() }}</div>
      }
      @if (sparkline().length > 0) {
        <svg class="metric-card__sparkline" viewBox="0 0 100 40" preserveAspectRatio="none">
          <defs>
            <linearGradient [attr.id]="'sparkGrad-' + sparkId" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#777EF0" stop-opacity="0.35"/>
              <stop offset="100%" stop-color="#777EF0" stop-opacity="0.02"/>
            </linearGradient>
            <filter [attr.id]="'sparkGlow-' + sparkId">
              <feGaussianBlur stdDeviation="2" result="blur"/>
              <feMerge>
                <feMergeNode in="blur"/>
                <feMergeNode in="SourceGraphic"/>
              </feMerge>
            </filter>
          </defs>
          <polyline [attr.points]="sparklineFill()" [attr.fill]="'url(#sparkGrad-' + sparkId + ')'" stroke="none" />
          <polyline [attr.points]="sparklineStroke()" fill="none" stroke="#777EF0" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" [attr.filter]="'url(#sparkGlow-' + sparkId + ')'" />
        </svg>
      }
    </div>
  `,
  styles: [`
    :host {
      display: block;
      height: 100%;
    }
    .metric-card {
      background: rgba(81, 82, 162, 0.2);
      backdrop-filter: blur(8px);
      -webkit-backdrop-filter: blur(8px);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: var(--radius-lg);
      padding: 20px;
      position: relative;
      overflow: hidden;
      transition: all 0.3s ease-out;
      box-shadow:
        0 4px 24px rgba(30, 30, 80, 0.4),
        0 1px 0 rgba(255, 255, 255, 0.03) inset;
      height: 100%;
      display: flex;
      flex-direction: column;
      min-height: 140px;
    }
    .metric-card:hover {
      border-color: rgba(255, 255, 255, 0.14);
      transform: translateY(-4px);
      box-shadow:
        0 16px 48px rgba(20, 20, 60, 0.5),
        0 0 20px rgba(119, 126, 240, 0.08),
        0 1px 0 rgba(255, 255, 255, 0.05) inset;
    }
    .metric-card__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
    }
    .metric-card__label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      opacity: 0.65;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      font-weight: 500;
    }
    .metric-card__trend {
      font-size: 0.85rem;
      font-weight: 700;
    }
    .metric-card__trend.up { color: var(--healthy); }
    .metric-card__trend.down { color: var(--failed); }
    .metric-card__value {
      font-family: var(--font-mono);
      font-size: 1.75rem;
      font-weight: 800;
      color: var(--rw-white);
      line-height: 1.2;
      margin-bottom: 4px;
      letter-spacing: -0.02em;
    }
    .metric-card__subtitle {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      opacity: 0.6;
    }
    .metric-card__sparkline {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 40px;
    }
  `]
})
export class MetricCardComponent implements AfterViewInit, OnDestroy {
  label = input<string>('');
  displayValue = input<string>('0');
  subtitle = input<string>('');
  trend = input<'up' | 'down' | 'flat' | ''>('');
  sparkline = input<number[]>([]);

  card = viewChild<ElementRef>('card');
  private ctx: gsap.Context | undefined;

  /* Unique ID for SVG gradient/filter refs to avoid collisions */
  sparkId = Math.random().toString(36).substring(2, 8);

  ngAfterViewInit() {
    const el = this.card();
    if (el) {
      this.ctx = gsap.context(() => {
        gsap.from(el.nativeElement, {
          y: 24,
          opacity: 0,
          duration: 0.6,
          ease: 'power2.out'
        });
      });
    }
  }

  ngOnDestroy() {
    this.ctx?.revert();
  }

  /** Stroke-only points (the visible line) */
  sparklineStroke(): string {
    const data = this.sparkline();
    if (data.length < 2) return '';
    const max = Math.max(...data);
    const min = Math.min(...data);
    const range = max - min || 1;
    return data.map((v, i) => {
      const x = (i / (data.length - 1)) * 100;
      const y = 40 - ((v - min) / range) * 36;
      return `${x},${y}`;
    }).join(' ');
  }

  /** Closed polygon for the gradient fill area */
  sparklineFill(): string {
    const stroke = this.sparklineStroke();
    if (!stroke) return '';
    return stroke + ' 100,40 0,40';
  }
}
