import { Component, input } from '@angular/core';
import { HealthStatus } from '../../core/models/types';

@Component({
  selector: 'rw-status-pill',
  standalone: true,
  template: `
    <span class="pill" [class]="'pill--' + status()">
      <span class="pill__dot-wrap">
        <span class="pill__dot"></span>
        <span class="pill__ping"></span>
      </span>
      <span class="pill__label">{{ label() || status() }}</span>
    </span>
  `,
  styles: [`
    .pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      border-radius: 100px;
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      white-space: nowrap;
      transition: all 0.3s ease-out;
    }
    .pill__dot-wrap {
      position: relative;
      width: 6px;
      height: 6px;
      flex-shrink: 0;
    }
    .pill__dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      position: relative;
      z-index: 1;
    }
    .pill__ping {
      position: absolute;
      inset: 0;
      border-radius: 50%;
      animation: rw-ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
    }

    /* Healthy — green inner glow + ping */
    .pill--healthy {
      background: rgba(110, 231, 183, 0.1);
      color: var(--healthy);
      box-shadow: inset 0 0 8px rgba(110, 231, 183, 0.12);
    }
    .pill--healthy .pill__dot {
      background: var(--healthy);
      box-shadow: 0 0 6px rgba(110, 231, 183, 0.5);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .pill--healthy .pill__ping { background: var(--healthy); }

    /* Degraded — amber inner glow + ping */
    .pill--degraded {
      background: rgba(252, 211, 77, 0.1);
      color: var(--degraded);
      box-shadow: inset 0 0 8px rgba(252, 211, 77, 0.12);
    }
    .pill--degraded .pill__dot {
      background: var(--degraded);
      box-shadow: 0 0 6px rgba(252, 211, 77, 0.5);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .pill--degraded .pill__ping { background: var(--degraded); }

    /* Failed — red inner glow + ping */
    .pill--failed {
      background: rgba(248, 113, 113, 0.1);
      color: var(--failed);
      box-shadow: inset 0 0 8px rgba(248, 113, 113, 0.15);
    }
    .pill--failed .pill__dot {
      background: var(--failed);
      box-shadow: 0 0 6px rgba(248, 113, 113, 0.5);
      animation: rw-breathe 2s ease-in-out infinite;
    }
    .pill--failed .pill__ping { background: var(--failed); }

    /* Pending / Building / In Progress — accent ping */
    .pill--pending, .pill--building, .pill--in_progress {
      background: rgba(119, 126, 240, 0.1);
      color: var(--rw-accent);
      box-shadow: inset 0 0 8px rgba(119, 126, 240, 0.12);
    }
    .pill--pending .pill__dot, .pill--building .pill__dot, .pill--in_progress .pill__dot {
      background: var(--rw-accent);
      box-shadow: 0 0 6px rgba(119, 126, 240, 0.5);
      animation: rw-breathe 2s ease-in-out infinite;
    }
    .pill--pending .pill__ping, .pill--building .pill__ping, .pill--in_progress .pill__ping {
      background: var(--rw-accent);
    }

    /* Complete — green inner glow + ping */
    .pill--complete {
      background: rgba(110, 231, 183, 0.1);
      color: var(--healthy);
      box-shadow: inset 0 0 8px rgba(110, 231, 183, 0.12);
    }
    .pill--complete .pill__dot {
      background: var(--healthy);
      box-shadow: 0 0 6px rgba(110, 231, 183, 0.4);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .pill--complete .pill__ping { background: var(--healthy); }

    /* Rolled Back — amber inner glow + ping */
    .pill--rolled_back {
      background: rgba(252, 211, 77, 0.1);
      color: var(--degraded);
      box-shadow: inset 0 0 8px rgba(252, 211, 77, 0.1);
    }
    .pill--rolled_back .pill__dot {
      background: var(--degraded);
      box-shadow: 0 0 6px rgba(252, 211, 77, 0.4);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .pill--rolled_back .pill__ping { background: var(--degraded); }
  `]
})
export class StatusPillComponent {
  status = input<string>('healthy');
  label = input<string>('');
}
