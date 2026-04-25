import { Component, inject, computed } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';
import { LogStreamComponent } from '../../shared/log-stream/log-stream.component';

@Component({
  selector: 'rw-builds',
  standalone: true,
  imports: [StatusPillComponent, LogStreamComponent],
  template: `
    <div class="builds">
      <h1 class="page-title">Build Pipeline</h1>

      <!-- Active Build -->
      @if (activeBuild(); as build) {
        <div class="active-build rw-card-static">
          <div class="active-build__header">
            <div class="active-build__info">
              <span class="active-build__name">{{ build.workloadName }}</span>
              <span class="sha-badge">{{ build.commitSha }}</span>
              <span class="branch-badge">{{ build.branch }}</span>
              <span class="trigger-badge">{{ build.trigger }}</span>
            </div>
            <rw-status-pill [status]="build.status" [label]="build.status" />
          </div>

          <!-- Step Timeline -->
          <div class="steps-timeline">
            @for (step of build.steps; track step.name) {
              <div class="step" [class]="'step--' + step.status">
                <div class="step__icon material-symbols-outlined">{{ step.icon }}</div>
                <div class="step__label">{{ step.name }}</div>
                <div class="step__time">{{ step.elapsed }}</div>
                <div class="step__dot"></div>
              </div>
              @if (!$last) {
                <div class="step__connector" [class.active]="step.status === 'complete'"></div>
              }
            }
          </div>

          <!-- Live Logs -->
          <rw-log-stream [title]="'Build Logs — ' + build.workloadName" [lines]="mockData.activeBuildLogs()" />
        </div>
      }

      <!-- Build History -->
      <div class="section-header">
        <h2 class="section-title">Build History</h2>
      </div>
      <div class="history-table">
        <div class="history-header">
          <span>Build ID</span><span>Workload</span><span>Branch</span><span>Status</span><span>Duration</span><span>Digest</span><span>Time</span>
        </div>
        @for (build of allBuilds(); track build.id) {
          <div class="history-row">
            <span class="mono">{{ build.id }}</span>
            <span class="history-row__name">{{ build.workloadName }}</span>
            <span class="mono">{{ build.branch }}</span>
            <rw-status-pill [status]="build.status" [label]="build.status" />
            <span class="mono">{{ build.duration }}</span>
            <span class="mono digest">{{ build.imageDigest || '—' }}</span>
            <span class="mono time">{{ timeAgo(build.timestamp) }}</span>
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .builds { max-width: 1400px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }
    .active-build { margin-bottom: 32px; }
    .active-build__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
    }
    .active-build__info { display: flex; align-items: center; gap: 10px; }
    .active-build__name {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
      font-weight: 600;
    }
    .sha-badge {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 2px 8px;
      background: rgba(119, 126, 240, 0.12);
      color: var(--rw-accent);
      border-radius: 4px;
    }
    .branch-badge {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 2px 8px;
      background: rgba(81, 82, 162, 0.4);
      color: var(--rw-mist);
      border-radius: 4px;
    }
    .trigger-badge {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 2px 8px;
      background: rgba(110, 231, 183, 0.1);
      color: var(--healthy);
      border-radius: 4px;
      text-transform: uppercase;
    }

    /* Steps Timeline */
    .steps-timeline {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0;
      margin-bottom: 20px;
      padding: 16px;
      background: rgba(66, 68, 127, 0.3);
      border-radius: var(--radius-md);
    }
    .step {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 4px;
      min-width: 80px;
      position: relative;
    }
    .step__icon { font-size: 1.3rem; }
    .step__label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mist);
    }
    .step__time {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .step__dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      margin-top: 4px;
    }
    .step--complete .step__dot { background: var(--healthy); }
    .step--running .step__dot {
      background: var(--rw-accent);
      animation: rw-pulse-dot 1.5s ease-in-out infinite;
    }
    .step--pending .step__dot { background: var(--rw-mid); opacity: 0.4; }
    .step--failed .step__dot { background: var(--failed); }
    .step__connector {
      width: 40px;
      height: 2px;
      background: var(--rw-mid);
      opacity: 0.3;
      margin: 0 4px;
      align-self: center;
      margin-top: 28px;
    }
    .step__connector.active {
      background: var(--healthy);
      opacity: 1;
    }

    /* History Table */
    .section-header { margin-bottom: 16px; }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
    }
    .history-table {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      overflow: hidden;
    }
    .history-header {
      display: grid;
      grid-template-columns: 1.5fr 1.5fr 1fr 1fr 1fr 2fr 1fr;
      gap: 8px;
      padding: 10px 16px;
      background: rgba(66, 68, 127, 0.5);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .history-row {
      display: grid;
      grid-template-columns: 1.5fr 1.5fr 1fr 1fr 1fr 2fr 1fr;
      gap: 8px;
      padding: 10px 16px;
      align-items: center;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      transition: background var(--transition-fast);
    }
    .history-row:hover { background: rgba(91, 106, 189, 0.08); }
    .history-row__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      color: var(--rw-white);
      font-weight: 500;
    }
    .mono { font-family: var(--font-mono); font-size: var(--text-xs); color: var(--rw-mist); }
    .digest { color: var(--rw-mid); overflow: hidden; text-overflow: ellipsis; }
    .time { color: var(--rw-mid); }

    @media (max-width: 1024px) {
      .steps-timeline { flex-wrap: wrap; }
      .history-header, .history-row { grid-template-columns: 1fr 1fr 1fr; }
    }
  `]
})
export class BuildsComponent {
  mockData = inject(MockDataService);

  activeBuild = computed(() => {
    const builds = this.mockData.builds();
    return builds.find(b => b.status === 'building') || builds[0];
  });

  allBuilds = this.mockData.builds;

  timeAgo(ts: string): string {
    const diff = Date.now() - new Date(ts).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'now';
    if (mins < 60) return `${mins}m`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h`;
    return `${Math.floor(hours / 24)}d`;
  }
}
