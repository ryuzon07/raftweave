import { Component, inject, computed, AfterViewInit, OnDestroy, ElementRef, viewChild } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { MetricCardComponent } from '../../shared/metric-card/metric-card.component';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';
import { AnimationService } from '../../core/services/animation.service';
import { gsap } from 'gsap';

@Component({
  selector: 'rw-dashboard',
  standalone: true,
  imports: [MetricCardComponent, StatusPillComponent],
  template: `
    <div class="dashboard" #dashboardEl>
      <h1 class="page-title">Overview</h1>

      <!-- Hero Stat Row -->
      <div class="stat-row">
        <rw-metric-card
          label="Active Workloads"
          [displayValue]="overview().workloadCount.toString()"
          trend="up"
          [sparkline]="[8, 9, 10, 11, 10, 12, 12]"
        />
        <rw-metric-card
          label="Cluster Health"
          [displayValue]="overview().clusterHealthPct.toFixed(1) + '%'"
          subtitle="6 nodes across 3 clouds"
          trend="flat"
          [sparkline]="[99.5, 99.7, 99.8, 99.6, 99.7, 99.9, 99.7]"
        />
        <rw-metric-card
          label="Avg Replication Lag"
          [displayValue]="overview().avgLagMs.toFixed(0) + 'ms'"
          subtitle="Target: < 500ms"
          [trend]="overview().avgLagMs > 500 ? 'down' : 'up'"
          [sparkline]="[250, 300, 280, 340, 310, 290, 342]"
        />
        <rw-metric-card
          label="Last Failover"
          [displayValue]="lastFailoverText()"
          [subtitle]="lastFailoverRto()"
          [sparkline]="[]"
        />
      </div>

      <!-- Two Column Layout -->
      <div class="dashboard-split">
        <!-- Cluster Mini-Map -->
        <div class="dashboard-split__main">
          <div class="section-header">
            <h2 class="section-title">Cluster Topology</h2>
            <rw-status-pill [status]="clusterHealth()" [label]="clusterHealth()" />
          </div>
          <div class="cluster-mini">
            @for (cloud of cloudGroups(); track cloud.name) {
              <div class="cloud-zone" [class]="'cloud-zone--' + cloud.name.toLowerCase()">
                <div class="cloud-zone__header">
                  <span class="cloud-zone__icon material-symbols-outlined" [style.color]="cloud.color">cloud</span>
                  <span class="cloud-zone__name">{{ cloud.name }}</span>
                </div>
                @for (node of cloud.nodes; track node.id) {
                  <div class="cloud-zone__node" [class.leader]="node.id === topology().leaderId">
                    <span class="node-dot" [class]="'node-dot--' + node.health"></span>
                    <span class="node-region">{{ node.region }}</span>
                    @if (node.id === topology().leaderId) {
                      <span class="node-crown material-symbols-outlined">workspace_premium</span>
                    }
                    <span class="node-lag">{{ node.lagMs.toFixed(0) }}ms</span>
                  </div>
                }
              </div>
            }
          </div>
        </div>

        <!-- Failover Feed -->
        <div class="dashboard-split__side">
          <div class="section-header">
            <h2 class="section-title">Recent Failovers</h2>
          </div>
          <div class="failover-feed">
            @for (event of recentFailovers(); track event.eventId) {
              <div class="failover-item">
                <div class="failover-item__time">{{ timeAgo(event.occurredAt) }}</div>
                <div class="failover-item__detail">
                  <span class="failover-item__trigger" [class]="'trigger--' + event.trigger">
                    {{ formatTrigger(event.trigger) }}
                  </span>
                  <span class="failover-item__route">
                    {{ event.fromRegion }} → {{ event.toRegion }}
                  </span>
                </div>
                <div class="failover-item__meta">
                  <span>RTO: {{ event.rtoSeconds }}s</span>
                  <rw-status-pill [status]="event.status" [label]="event.status" />
                </div>
              </div>
            } @empty {
              <div class="empty-state">No recent failover events</div>
            }
          </div>
        </div>
      </div>

      <!-- Build Activity Feed -->
      <div class="section-header" style="margin-top: 24px;">
        <h2 class="section-title">Recent Builds</h2>
      </div>
      <div class="build-feed">
        @for (build of recentBuilds(); track build.id) {
          <div class="build-item">
            <rw-status-pill [status]="build.status" [label]="build.status" />
            <span class="build-item__name">{{ build.workloadName }}</span>
            <span class="build-item__sha mono">{{ build.commitSha }}</span>
            <span class="build-item__branch">{{ build.branch }}</span>
            <span class="build-item__duration mono">{{ build.duration }}</span>
            <span class="build-item__time">{{ timeAgo(build.timestamp) }}</span>
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .dashboard { max-width: 1400px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }
    .stat-row {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-bottom: 24px;
    }
    .dashboard-split {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: 24px;
    }
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
    }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
    }

    /* Cluster Mini Map */
    .cluster-mini {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
    .cloud-zone {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      padding: 14px;
    }
    .cloud-zone__header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 10px;
      padding-bottom: 8px;
      border-bottom: 1px solid rgba(91, 106, 189, 0.1);
    }
    .cloud-zone__icon { font-size: 1.1rem; }
    .cloud-zone__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      font-weight: 600;
      color: var(--rw-white);
    }
    .cloud-zone__node {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 6px 0;
      font-size: var(--text-xs);
    }
    .cloud-zone__node.leader {
      background: rgba(119, 126, 240, 0.1);
      margin: 0 -8px;
      padding: 6px 8px;
      border-radius: var(--radius-sm);
    }
    .node-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      flex-shrink: 0;
    }
    .node-dot--healthy { background: var(--healthy); }
    .node-dot--degraded { background: var(--degraded); }
    .node-dot--failed { background: var(--failed); }
    .node-region {
      font-family: var(--font-mono);
      color: var(--rw-mist);
      flex: 1;
    }
    .node-crown { font-size: 0.8rem; }
    .node-lag {
      font-family: var(--font-mono);
      color: var(--rw-mid);
    }

    /* Failover Feed */
    .failover-feed {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      padding: 4px;
      max-height: 400px;
      overflow-y: auto;
    }
    .failover-item {
      padding: 12px;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      transition: background var(--transition-fast);
    }
    .failover-item:hover { background: rgba(91, 106, 189, 0.08); }
    .failover-item:last-child { border-bottom: none; }
    .failover-item__time {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      margin-bottom: 4px;
    }
    .failover-item__detail {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 4px;
    }
    .failover-item__trigger {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 1px 6px;
      border-radius: 4px;
    }
    .trigger--node_failure { background: rgba(248, 113, 113, 0.15); color: var(--failed); }
    .trigger--network_partition { background: rgba(252, 211, 77, 0.15); color: var(--degraded); }
    .trigger--manual { background: rgba(119, 126, 240, 0.15); color: var(--rw-accent); }
    .failover-item__route {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
    }
    .failover-item__meta {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }

    /* Build Feed */
    .build-feed {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      overflow: hidden;
    }
    .build-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 16px;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      transition: background var(--transition-fast);
    }
    .build-item:hover { background: rgba(91, 106, 189, 0.08); }
    .build-item:last-child { border-bottom: none; }
    .build-item__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      color: var(--rw-white);
      font-weight: 500;
      min-width: 140px;
    }
    .build-item__sha {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-accent);
      background: rgba(119, 126, 240, 0.1);
      padding: 1px 6px;
      border-radius: 4px;
    }
    .build-item__branch {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
    }
    .build-item__duration {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      margin-left: auto;
    }
    .build-item__time {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .empty-state {
      padding: 24px;
      text-align: center;
      color: var(--rw-mid);
      font-size: var(--text-body);
    }
    .mono { font-family: var(--font-mono); }

    @media (max-width: 1024px) {
      .stat-row { grid-template-columns: repeat(2, 1fr); }
      .dashboard-split { grid-template-columns: 1fr; }
      .cluster-mini { grid-template-columns: 1fr; }
    }
    @media (max-width: 768px) {
      .stat-row { grid-template-columns: 1fr; }
    }
  `]
})
export class DashboardComponent implements AfterViewInit, OnDestroy {
  mockData = inject(MockDataService);
  private animService = inject(AnimationService);
  private ctx: gsap.Context | undefined;

  dashboardEl = viewChild<ElementRef>('dashboardEl');

  overview = this.mockData.workspaceOverview;
  topology = this.mockData.clusterTopology;

  cloudGroups = computed(() => {
    const nodes = this.topology().nodes;
    return [
      { name: 'AWS', color: '#F59E0B', nodes: nodes.filter(n => n.cloud === 'AWS') },
      { name: 'Azure', color: '#3B82F6', nodes: nodes.filter(n => n.cloud === 'Azure') },
      { name: 'GCP', color: '#EF4444', nodes: nodes.filter(n => n.cloud === 'GCP') },
    ];
  });

  clusterHealth = computed(() => {
    const nodes = this.topology().nodes;
    const unhealthy = nodes.filter(n => n.health !== 'healthy').length;
    if (unhealthy === 0) return 'healthy';
    if (unhealthy <= 2) return 'degraded';
    return 'failed';
  });

  recentFailovers = computed(() => this.mockData.failoverEvents().slice(0, 5));
  recentBuilds = computed(() => this.mockData.builds());

  ngAfterViewInit() {
    const el = this.dashboardEl();
    if (el) {
      this.ctx = gsap.context(() => {
        this.animService.fadeUpStagger('.stat-row rw-metric-card');
      }, el.nativeElement);
    }
  }

  ngOnDestroy() {
    this.ctx?.revert();
  }

  lastFailoverText(): string {
    const fo = this.overview().lastFailover;
    if (!fo) return 'None (30d)';
    return this.timeAgo(fo.occurredAt);
  }

  lastFailoverRto(): string {
    const fo = this.overview().lastFailover;
    if (!fo) return 'No incidents';
    return `RTO: ${fo.rtoSeconds}s`;
  }

  timeAgo(ts: string): string {
    const diff = Date.now() - new Date(ts).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  }

  formatTrigger(t: string): string {
    return t.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }
}
