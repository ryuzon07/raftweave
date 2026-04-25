import { Component, inject, signal, computed } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';
import { CloudBadgeComponent } from '../../shared/cloud-badge/cloud-badge.component';
import { DrawerComponent } from '../../shared/drawer/drawer.component';
import { Workload } from '../../core/models/types';

@Component({
  selector: 'rw-workloads',
  standalone: true,
  imports: [StatusPillComponent, CloudBadgeComponent, DrawerComponent],
  template: `
    <div class="workloads">
      <div class="page-header">
        <h1 class="page-title">Workloads</h1>
        <div class="page-actions">
          <button class="view-toggle material-symbols-outlined" (click)="viewMode.set(viewMode() === 'grid' ? 'list' : 'grid')">
            {{ viewMode() === 'grid' ? 'menu' : 'grid_view' }}
          </button>
          <a href="/onboarding" class="rw-btn-primary">+ Deploy New</a>
        </div>
      </div>

      @if (viewMode() === 'grid') {
        <div class="workload-grid">
          @for (wl of workloads(); track wl.id) {
            <div class="workload-card rw-card" (click)="selectWorkload(wl)">
              <div class="workload-card__header">
                <span class="workload-card__name">{{ wl.name }}</span>
                <rw-status-pill [status]="wl.health" />
              </div>
              <div class="workload-card__lang">
                <span class="lang-badge">{{ wl.language }}</span>
              </div>
              <div class="workload-card__regions">
                @for (r of wl.regions; track r.region) {
                  <rw-cloud-badge [cloud]="r.cloud" [region]="r.region" />
                }
              </div>
              <div class="workload-card__footer">
                <span class="workload-card__deploy">{{ wl.lastDeploy }}</span>
                <div class="workload-card__actions">
                  <button class="action-btn material-symbols-outlined" title="Redeploy">refresh</button>
                  <button class="action-btn material-symbols-outlined" title="Scale">swap_vert</button>
                  <button class="action-btn material-symbols-outlined" title="Logs">description</button>
                </div>
              </div>
            </div>
          }
        </div>
      } @else {
        <div class="workload-table">
          <div class="table-header">
            <span>Name</span><span>Language</span><span>Regions</span><span>Health</span><span>Last Deploy</span><span>Actions</span>
          </div>
          @for (wl of workloads(); track wl.id) {
            <div class="table-row" (click)="selectWorkload(wl)">
              <span class="table-row__name">{{ wl.name }}</span>
              <span class="lang-badge">{{ wl.language }}</span>
              <span class="table-row__regions">
                @for (r of wl.regions; track r.region) {
                  <rw-cloud-badge [cloud]="r.cloud" [region]="r.region" />
                }
              </span>
              <rw-status-pill [status]="wl.health" />
              <span class="table-row__time">{{ wl.lastDeploy }}</span>
              <div class="table-row__actions">
                <button class="action-btn material-symbols-outlined" title="Redeploy">refresh</button>
                <button class="action-btn material-symbols-outlined" title="Scale">swap_vert</button>
                <button class="action-btn material-symbols-outlined" title="Logs">description</button>
              </div>
            </div>
          }
        </div>
      }
    </div>

    <!-- Workload Detail Drawer -->
    <rw-drawer [open]="drawerOpen()" [title]="selectedWorkload()?.name || ''" (closed)="closeDrawer()">
      @if (selectedWorkload(); as wl) {
        <div class="drawer-section">
          <h4 class="drawer-section__title">Descriptor</h4>
          <pre class="rw-terminal">{{ wl.descriptorYaml }}</pre>
        </div>
        <div class="drawer-section">
          <h4 class="drawer-section__title">Deployment Regions</h4>
          <div class="drawer-regions">
            @for (r of wl.regions; track r.region) {
              <div class="drawer-region">
                <rw-cloud-badge [cloud]="r.cloud" [region]="r.region" />
                <rw-status-pill [status]="r.status" [label]="r.status" />
              </div>
            }
          </div>
        </div>
        <div class="drawer-section">
          <h4 class="drawer-section__title">Environment Variables</h4>
          <div class="env-vars">
            @for (env of wl.envVars; track env.key) {
              <div class="env-var">
                <span class="env-var__key">{{ env.key }}</span>
                <span class="env-var__val">{{ env.masked ? '••••••••' : env.value }}</span>
              </div>
            } @empty {
              <p class="text-muted">No environment variables configured</p>
            }
          </div>
        </div>
      }
    </rw-drawer>
  `,
  styles: [`
    .workloads { max-width: 1400px; margin: 0 auto; }
    .page-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 24px;
    }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
    }
    .page-actions { display: flex; align-items: center; gap: 12px; }
    .view-toggle {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.2);
      color: var(--rw-mist);
      width: 36px;
      height: 36px;
      border-radius: var(--radius-sm);
      font-size: 1.1rem;
      cursor: pointer;
      transition: all var(--transition-fast);
    }
    .view-toggle:hover { border-color: var(--rw-accent); color: var(--rw-white); }

    .workload-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 16px;
    }
    .workload-card { cursor: pointer; }
    .workload-card__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;
    }
    .workload-card__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      font-weight: 600;
      color: var(--rw-white);
    }
    .lang-badge {
      display: inline-block;
      padding: 2px 8px;
      background: rgba(119, 126, 240, 0.12);
      color: var(--rw-accent);
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      border-radius: 4px;
    }
    .workload-card__lang { margin-bottom: 10px; }
    .workload-card__regions {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      margin-bottom: 12px;
    }
    .workload-card__footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-top: 10px;
      border-top: 1px solid rgba(91, 106, 189, 0.1);
    }
    .workload-card__deploy {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .workload-card__actions { display: flex; gap: 4px; }
    .action-btn {
      background: transparent;
      border: 1px solid rgba(91, 106, 189, 0.2);
      color: var(--rw-mist);
      width: 28px;
      height: 28px;
      border-radius: var(--radius-sm);
      font-size: 0.8rem;
      cursor: pointer;
      transition: all var(--transition-fast);
    }
    .action-btn:hover {
      border-color: var(--rw-accent);
      color: var(--rw-white);
      background: rgba(119, 126, 240, 0.1);
    }

    /* Table view */
    .workload-table {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      overflow: hidden;
    }
    .table-header {
      display: grid;
      grid-template-columns: 2fr 1fr 3fr 1fr 1fr 1fr;
      gap: 12px;
      padding: 10px 16px;
      background: rgba(66, 68, 127, 0.5);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .table-row {
      display: grid;
      grid-template-columns: 2fr 1fr 3fr 1fr 1fr 1fr;
      gap: 12px;
      padding: 10px 16px;
      align-items: center;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      cursor: pointer;
      transition: background var(--transition-fast);
    }
    .table-row:hover { background: rgba(91, 106, 189, 0.08); }
    .table-row__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      color: var(--rw-white);
      font-weight: 500;
    }
    .table-row__regions { display: flex; flex-wrap: wrap; gap: 4px; }
    .table-row__time {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .table-row__actions { display: flex; gap: 4px; }

    /* Drawer sections */
    .drawer-section { margin-bottom: 24px; }
    .drawer-section__title {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      color: var(--rw-white);
      margin-bottom: 10px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .drawer-regions { display: flex; flex-direction: column; gap: 8px; }
    .drawer-region {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: rgba(81, 82, 162, 0.3);
      border-radius: var(--radius-sm);
    }
    .env-vars { display: flex; flex-direction: column; gap: 6px; }
    .env-var {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: rgba(10, 10, 18, 0.5);
      border-radius: var(--radius-sm);
      font-family: var(--font-mono);
      font-size: var(--text-xs);
    }
    .env-var__key { color: var(--rw-accent); }
    .env-var__val { color: var(--rw-mid); }
    .text-muted { color: var(--rw-mid); font-size: var(--text-xs); }
  `]
})
export class WorkloadsComponent {
  mockData = inject(MockDataService);
  workloads = this.mockData.workloads;
  viewMode = signal<'grid' | 'list'>('grid');
  drawerOpen = signal(false);
  selectedWorkload = signal<Workload | null>(null);

  selectWorkload(wl: Workload): void {
    this.selectedWorkload.set(wl);
    this.drawerOpen.set(true);
  }

  closeDrawer(): void {
    this.drawerOpen.set(false);
  }
}
