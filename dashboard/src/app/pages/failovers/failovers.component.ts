import { Component, inject, computed } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';
import { MetricCardComponent } from '../../shared/metric-card/metric-card.component';

@Component({
  selector: 'rw-failovers',
  standalone: true,
  imports: [StatusPillComponent, MetricCardComponent],
  template: `
    <div class="failovers">
      <h1 class="page-title">Failover Events</h1>

      <!-- Summary Stats Bar -->
      <div class="summary-bar">
        <rw-metric-card label="Total Failovers (30d)" [displayValue]="totalFailovers().toString()"
          [sparkline]="failoverSparkline()" />
        <rw-metric-card label="Avg RTO" [displayValue]="avgRto().toFixed(1) + 's'" subtitle="Recovery Time Objective" />
        <rw-metric-card label="Avg RPO" [displayValue]="avgRpo().toFixed(0) + 'ms'" subtitle="Recovery Point Objective" />
        <rw-metric-card label="Success Rate" [displayValue]="successRate().toFixed(0) + '%'"
          [trend]="successRate() >= 95 ? 'up' : 'down'"
          [sparkline]="successSparkline()" />
      </div>

      <!-- Timeline -->
      <div class="timeline">
        @for (event of failoverEvents(); track event.eventId; let i = $index; let last = $last) {
          <div class="timeline-item">
            <div class="timeline-line">
              <div class="timeline-dot-wrap">
                <div class="timeline-dot" [class]="'dot--' + event.status"></div>
                <div class="timeline-ping" [class]="'ping--' + event.status"></div>
              </div>
              @if (!last) {
                <div class="timeline-connector"
                  [style.opacity]="1 - (i / failoverEvents().length) * 0.7"></div>
              }
            </div>
            <div class="timeline-card rw-card">
              <div class="timeline-card__header">
                <div class="timeline-card__time">
                  <span class="time-abs">{{ formatDate(event.occurredAt) }}</span>
                  <span class="time-rel">{{ timeAgo(event.occurredAt) }}</span>
                </div>
                <rw-status-pill [status]="event.status" [label]="event.status" />
              </div>
              <div class="timeline-card__body">
                <div class="trigger-badge" [class]="'trigger--' + event.trigger">
                  {{ formatTrigger(event.trigger) }}
                </div>
                <div class="route-arrow">
                  <span class="route-from">{{ event.fromRegion }}</span>
                  <span class="arrow material-symbols-outlined">arrow_forward</span>
                  <span class="route-to">{{ event.toRegion }}</span>
                </div>
              </div>
              <div class="timeline-card__metrics">
                <div class="tm-item">
                  <span class="tm-label">RTO</span>
                  <span class="tm-value">{{ event.rtoSeconds }}s</span>
                </div>
                <div class="tm-item">
                  <span class="tm-label">RPO</span>
                  <span class="tm-value">{{ event.rpoMs }}ms</span>
                </div>
                <div class="tm-item">
                  <span class="tm-label">Raft Term</span>
                  <span class="tm-value accent">{{ event.term }}</span>
                </div>
              </div>
            </div>
          </div>
        } @empty {
          <div class="empty-state">
            <div class="empty-icon material-symbols-outlined">electric_bolt</div>
            <p>No failover events recorded</p>
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .failovers { max-width: 900px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
      letter-spacing: -0.02em;
    }
    .summary-bar {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-bottom: 32px;
    }

    /* ━━━ Timeline ━━━ */
    .timeline {
      display: flex;
      flex-direction: column;
      gap: 0;
    }
    .timeline-item {
      display: flex;
      gap: 20px;
    }
    .timeline-line {
      display: flex;
      flex-direction: column;
      align-items: center;
      width: 20px;
      flex-shrink: 0;
    }

    /* Timeline dots with neon glow + ping ring */
    .timeline-dot-wrap {
      position: relative;
      width: 12px;
      height: 12px;
      flex-shrink: 0;
      margin-top: 20px;
    }
    .timeline-dot {
      width: 12px;
      height: 12px;
      border-radius: 50%;
      position: relative;
      z-index: 1;
      transition: all 0.3s ease-out;
    }
    .timeline-ping {
      position: absolute;
      inset: 0;
      border-radius: 50%;
      animation: rw-ping 2.5s cubic-bezier(0, 0, 0.2, 1) infinite;
    }
    .dot--complete {
      background: var(--healthy);
      box-shadow: 0 0 8px rgba(110, 231, 183, 0.4);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .ping--complete { background: var(--healthy); }
    .dot--in_progress {
      background: var(--rw-accent);
      box-shadow: 0 0 8px rgba(119, 126, 240, 0.5);
      animation: rw-breathe 2s ease-in-out infinite;
    }
    .ping--in_progress { background: var(--rw-accent); }
    .dot--rolled_back {
      background: var(--degraded);
      box-shadow: 0 0 8px rgba(252, 211, 77, 0.4);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .ping--rolled_back { background: var(--degraded); }

    /* Gradient timeline connector — bright at top, fading down */
    .timeline-connector {
      width: 2px;
      flex: 1;
      background: linear-gradient(to bottom, rgba(119, 126, 240, 0.4), rgba(119, 126, 240, 0.05));
      margin: 4px 0;
      border-radius: 1px;
    }

    /* ━━━ Card Overrides for Timeline ━━━ */
    .timeline-card {
      flex: 1;
      margin-bottom: 12px;
    }
    .timeline-card__header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 12px;
    }
    .time-abs {
      font-family: var(--font-mono);
      font-size: var(--text-body);
      color: var(--rw-white);
      display: block;
      font-weight: 500;
    }
    .time-rel {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      opacity: 0.7;
    }
    .timeline-card__body {
      display: flex;
      align-items: center;
      gap: 16px;
      margin-bottom: 12px;
    }
    .trigger-badge {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 3px 8px;
      border-radius: 4px;
      text-transform: uppercase;
      font-weight: 500;
      letter-spacing: 0.03em;
    }
    .trigger--node_failure {
      background: rgba(248, 113, 113, 0.12);
      color: var(--failed);
      box-shadow: inset 0 0 6px rgba(248, 113, 113, 0.08);
    }
    .trigger--network_partition {
      background: rgba(252, 211, 77, 0.12);
      color: var(--degraded);
      box-shadow: inset 0 0 6px rgba(252, 211, 77, 0.08);
    }
    .trigger--manual {
      background: rgba(119, 126, 240, 0.12);
      color: var(--rw-accent);
      box-shadow: inset 0 0 6px rgba(119, 126, 240, 0.08);
    }
    .route-arrow {
      display: flex;
      align-items: center;
      gap: 8px;
      font-family: var(--font-mono);
      font-size: var(--text-body);
    }
    .route-from { color: var(--failed); font-weight: 500; }
    .arrow {
      color: var(--rw-mid);
      font-size: 1rem;
      opacity: 0.6;
    }
    .route-to { color: var(--healthy); font-weight: 500; }

    /* ━━━ Metrics Row ━━━ */
    .timeline-card__metrics {
      display: flex;
      gap: 20px;
      padding-top: 12px;
      border-top: 1px solid rgba(255, 255, 255, 0.06);
    }
    .tm-item { display: flex; flex-direction: column; gap: 2px; }
    .tm-label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      opacity: 0.6;
      letter-spacing: 0.04em;
    }
    .tm-value {
      font-family: var(--font-mono);
      font-size: var(--text-body);
      color: var(--rw-white);
      font-weight: 700;
    }
    .tm-value.accent { color: var(--rw-accent); }

    .empty-state {
      text-align: center;
      padding: 60px 20px;
      color: var(--rw-mid);
    }
    .empty-icon { font-size: 3rem; margin-bottom: 12px; }

    @media (max-width: 768px) {
      .summary-bar { grid-template-columns: repeat(2, 1fr); }
    }
  `]
})
export class FailoversComponent {
  mockData = inject(MockDataService);
  failoverEvents = this.mockData.failoverEvents;

  totalFailovers = computed(() => this.failoverEvents().length);

  avgRto = computed(() => {
    const events = this.failoverEvents();
    if (events.length === 0) return 0;
    return events.reduce((sum, e) => sum + e.rtoSeconds, 0) / events.length;
  });

  avgRpo = computed(() => {
    const events = this.failoverEvents();
    if (events.length === 0) return 0;
    return events.reduce((sum, e) => sum + e.rpoMs, 0) / events.length;
  });

  successRate = computed(() => {
    const events = this.failoverEvents();
    if (events.length === 0) return 100;
    const successful = events.filter(e => e.status === 'complete').length;
    return (successful / events.length) * 100;
  });

  /* Sparkline data for metric cards */
  failoverSparkline = computed(() => {
    const events = this.failoverEvents();
    return events.slice(0, 7).map((_, i) => Math.max(1, events.length - i + Math.floor(Math.random() * 2)));
  });

  successSparkline = computed(() => {
    return [92, 95, 88, 96, 100, 94, 97];
  });

  formatDate(ts: string): string {
    try {
      return new Date(ts).toLocaleString('en-US', {
        month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false
      });
    } catch { return '—'; }
  }

  timeAgo(ts: string): string {
    const diff = Date.now() - new Date(ts).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  }

  formatTrigger(t: string): string {
    return t.replace(/_/g, ' ');
  }
}
