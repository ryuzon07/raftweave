import { Component, inject, computed, AfterViewInit, OnDestroy, ElementRef, viewChild, effect } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { MetricCardComponent } from '../../shared/metric-card/metric-card.component';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';

@Component({
  selector: 'rw-replication',
  standalone: true,
  imports: [MetricCardComponent],
  template: `
    <div class="replication">
      <h1 class="page-title">Replication</h1>

      <!-- Header Metrics -->
      <div class="metrics-row">
        <rw-metric-card label="Primary Region" [displayValue]="'us-east-1'" subtitle="AWS" />
        <rw-metric-card
          label="Max Lag Across Standbys"
          [displayValue]="maxLag().toFixed(0) + 'ms'"
          [trend]="maxLag() > 5000 ? 'down' : 'up'"
          [sparkline]="lagSparkline()"
        />
        <rw-metric-card label="RPO Status"
          [displayValue]="maxLag() > 5000 ? 'VIOLATED' : 'ENFORCED'"
          [subtitle]="'Threshold: 5000ms'"
        />
      </div>

      <!-- Lag Chart -->
      <div class="section-header">
        <h2 class="section-title">Replication Lag — Last 5 Minutes</h2>
      </div>
      <div class="lag-chart-container rw-card-static">
        <div class="chart-area" #chartArea>
          <svg class="chart-svg" [attr.viewBox]="'0 0 ' + chartWidth + ' ' + chartHeight" preserveAspectRatio="none">
            <!-- Threshold line -->
            <line x1="0" [attr.y1]="thresholdY()" [attr.x1]="chartWidth" [attr.y2]="thresholdY()"
              stroke="var(--failed)" stroke-width="1" stroke-dasharray="6 4" opacity="0.6" />
            <text [attr.x]="chartWidth - 5" [attr.y]="thresholdY() - 6"
              text-anchor="end" fill="var(--failed)" font-size="10" font-family="var(--font-mono)">
              5000ms RPO
            </text>

            <!-- Data lines -->
            @for (series of chartSeries(); track series.region) {
              <polyline
                [attr.points]="series.points"
                fill="none"
                [attr.stroke]="series.color"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            }

            <!-- Y-axis labels -->
            @for (label of yLabels; track label) {
              <text x="2" [attr.y]="labelY(label) - 4"
                fill="var(--rw-mid)" font-size="9" font-family="var(--font-mono)">
                {{ label }}ms
              </text>
            }
          </svg>

          <!-- Legend -->
          <div class="chart-legend">
            @for (series of chartSeries(); track series.region) {
              <div class="legend-item">
                <span class="legend-dot" [style.background]="series.color"></span>
                <span class="legend-label">{{ series.region }}</span>
                <span class="legend-value">{{ series.lastValue.toFixed(0) }}ms</span>
              </div>
            }
          </div>
        </div>
      </div>

      <!-- Replication Events Table -->
      <div class="section-header" style="margin-top: 24px;">
        <h2 class="section-title">Replication Events</h2>
      </div>
      <div class="events-table">
        <div class="et-header">
          <span>Timestamp</span><span>Region</span><span>Event Type</span><span>Lag</span><span>Action</span>
        </div>
        @for (event of replicationEvents(); track event.timestamp) {
          <div class="et-row">
            <span class="mono">{{ formatTime(event.timestamp) }}</span>
            <span class="mono">{{ event.region }}</span>
            <span class="event-type" [class]="'et--' + event.eventType.toLowerCase()">{{ event.eventType }}</span>
            <span class="mono" [class.lag-high]="event.lagAtEvent > 5000">{{ event.lagAtEvent }}ms</span>
            <span class="mono action">{{ event.actionTaken }}</span>
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .replication { max-width: 1400px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }
    .metrics-row {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 16px;
      margin-bottom: 24px;
    }
    .section-header { margin-bottom: 12px; }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
    }

    /* Chart */
    .lag-chart-container { padding: 16px; }
    .chart-area { position: relative; }
    .chart-svg {
      width: 100%;
      height: 220px;
      background: rgba(10, 10, 18, 0.3);
      border-radius: var(--radius-sm);
    }
    .chart-legend {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      padding: 12px 0 0;
    }
    .legend-item {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .legend-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
    }
    .legend-label {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
    }
    .legend-value {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }

    /* Events Table */
    .events-table {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      overflow: hidden;
    }
    .et-header {
      display: grid;
      grid-template-columns: 2fr 1fr 1.5fr 1fr 2fr;
      gap: 8px;
      padding: 10px 16px;
      background: rgba(66, 68, 127, 0.5);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .et-row {
      display: grid;
      grid-template-columns: 2fr 1fr 1.5fr 1fr 2fr;
      gap: 8px;
      padding: 10px 16px;
      align-items: center;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      transition: background var(--transition-fast);
    }
    .et-row:hover { background: rgba(91, 106, 189, 0.08); }
    .mono { font-family: var(--font-mono); font-size: var(--text-xs); color: var(--rw-mist); }
    .lag-high { color: var(--failed) !important; }
    .event-type {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 2px 6px;
      border-radius: 4px;
      display: inline-block;
      width: fit-content;
    }
    .et--lag_spike { background: rgba(248, 113, 113, 0.15); color: var(--failed); }
    .et--catchup_complete { background: rgba(110, 231, 183, 0.15); color: var(--healthy); }
    .et--sync_pause { background: rgba(252, 211, 77, 0.15); color: var(--degraded); }
    .action { color: var(--rw-mid); }

    @media (max-width: 1024px) {
      .metrics-row { grid-template-columns: 1fr; }
    }
  `]
})
export class ReplicationComponent {
  mockData = inject(MockDataService);
  replicationEvents = this.mockData.replicationEvents;

  chartWidth = 800;
  chartHeight = 200;
  maxY = 6000;
  yLabels = [0, 1000, 2000, 3000, 4000, 5000, 6000];

  regionColors: Record<string, string> = {
    'us-east-1': '#777EF0',
    'eastus': '#6EE7B7',
    'us-central1': '#FCD34D',
    'eu-west-1': '#F87171',
    'westeurope': '#5B6ABD'
  };

  maxLag = computed(() => {
    const lag = this.mockData.replicationLag();
    if (lag.length === 0) return 0;
    return Math.max(...lag.map(l => l.lagMs));
  });

  lagSparkline = computed(() => {
    const history = this.mockData.replicationHistory();
    return history.slice(-10).map(points =>
      Math.max(...points.map(p => p.lagMs))
    );
  });

  chartSeries = computed(() => {
    const history = this.mockData.replicationHistory();
    if (history.length === 0) return [];

    const regions = ['us-east-1', 'eastus', 'us-central1', 'eu-west-1', 'westeurope'];
    return regions.map(region => {
      const points = history.map((snapshot, i) => {
        const point = snapshot.find(p => p.region === region);
        const x = (i / Math.max(1, history.length - 1)) * this.chartWidth;
        const y = this.chartHeight - ((point?.lagMs || 0) / this.maxY) * this.chartHeight;
        return `${x},${Math.max(0, Math.min(this.chartHeight, y))}`;
      }).join(' ');

      const lastSnapshot = history[history.length - 1];
      const lastValue = lastSnapshot?.find(p => p.region === region)?.lagMs || 0;

      return {
        region,
        points,
        color: this.regionColors[region] || '#777EF0',
        lastValue
      };
    });
  });

  thresholdY(): number {
    return this.chartHeight - (5000 / this.maxY) * this.chartHeight;
  }

  labelY(ms: number): number {
    return this.chartHeight - (ms / this.maxY) * this.chartHeight;
  }

  formatTime(ts: string): string {
    try {
      return new Date(ts).toLocaleTimeString('en-US', { hour12: false });
    } catch { return '—'; }
  }
}
