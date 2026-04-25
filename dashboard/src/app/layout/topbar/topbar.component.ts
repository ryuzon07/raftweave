import { Component, inject, computed } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { filter, map } from 'rxjs';
import { MockDataService } from '../../core/services/mock-data.service';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'rw-topbar',
  standalone: true,
  template: `
    <header class="topbar">
      <div class="topbar__left">
        <span class="topbar__breadcrumb">
          @for (crumb of breadcrumbs(); track $index; let last = $last) {
            <span class="topbar__crumb" [class.active]="last">{{ crumb }}</span>
            @if (!last) {
              <span class="topbar__crumb-sep">/</span>
            }
          }
        </span>
      </div>

      <div class="topbar__right">
        <!-- Connection Status -->
        @if (mockData.connectionStatus() === 'reconnecting') {
          <span class="topbar__pill topbar__pill--warn">
            <span class="topbar__pill-dot warn"></span>
            Reconnecting...
          </span>
        }

        <!-- Healthy Nodes -->
        <span class="topbar__pill topbar__pill--status">
          <span class="topbar__pill-dot-wrap">
            <span class="topbar__pill-dot healthy"></span>
            <span class="topbar__pill-ping healthy"></span>
          </span>
          {{ mockData.healthyNodeCount() }} nodes healthy
        </span>

        <!-- Workspace name -->
        <span class="topbar__workspace">Production</span>

        <!-- User Avatar -->
        <div class="topbar__avatar" title="{{ auth.user()?.name || 'User' }}">
          {{ (auth.user()?.name || 'U')[0] }}
        </div>
      </div>
    </header>
  `,
  styles: [`
    .topbar {
      position: fixed;
      top: 0;
      left: var(--sidebar-width);
      right: 0;
      height: var(--topbar-height);
      background: rgba(56, 58, 110, 0.6);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border-bottom: 1px solid rgba(255, 255, 255, 0.06);
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
      z-index: 90;
      transition: left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    }
    .topbar__left {
      display: flex;
      align-items: center;
    }
    .topbar__breadcrumb {
      display: flex;
      align-items: center;
      gap: 6px;
      font-family: var(--font-body);
      font-size: var(--text-body);
    }
    .topbar__crumb {
      color: var(--rw-mid);
    }
    .topbar__crumb.active {
      color: var(--rw-white);
      font-weight: 500;
    }
    .topbar__crumb-sep {
      color: var(--rw-mid);
      opacity: 0.4;
    }
    .topbar__right {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .topbar__pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      border-radius: 100px;
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      border: 1px solid rgba(255, 255, 255, 0.06);
    }
    .topbar__pill--status {
      background: rgba(110, 231, 183, 0.06);
      color: var(--healthy);
    }
    .topbar__pill--warn {
      background: rgba(252, 211, 77, 0.06);
      color: var(--degraded);
      animation: rw-pulse 1.5s ease-in-out infinite;
    }
    .topbar__pill-dot-wrap {
      position: relative;
      width: 6px;
      height: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .topbar__pill-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      position: relative;
      z-index: 1;
    }
    .topbar__pill-ping {
      position: absolute;
      inset: 0;
      border-radius: 50%;
      animation: rw-ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
    }
    .topbar__pill-dot.healthy {
      background: var(--healthy);
      box-shadow: 0 0 6px rgba(110, 231, 183, 0.5);
      animation: rw-breathe 3s ease-in-out infinite;
    }
    .topbar__pill-ping.healthy {
      background: var(--healthy);
    }
    .topbar__pill-dot.warn {
      background: var(--degraded);
      box-shadow: 0 0 6px rgba(252, 211, 77, 0.5);
      animation: rw-pulse-dot 1s ease-in-out infinite;
    }
    .topbar__workspace {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      padding: 3px 10px;
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: var(--radius-sm);
    }
    .topbar__avatar {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      background: var(--rw-accent);
      color: var(--rw-white);
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--font-heading);
      font-size: 0.75rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.3s ease-out;
    }
    .topbar__avatar:hover {
      transform: scale(1.1);
      box-shadow: 0 0 16px rgba(119, 126, 240, 0.5);
    }
    @media (max-width: 768px) {
      .topbar {
        left: 0;
      }
      .topbar__workspace { display: none; }
    }
  `]
})
export class TopbarComponent {
  mockData = inject(MockDataService);
  auth = inject(AuthService);
  private router = inject(Router);

  private routeUrl = toSignal(
    this.router.events.pipe(
      filter(e => e instanceof NavigationEnd),
      map((e: NavigationEnd) => e.urlAfterRedirects)
    ),
    { initialValue: '/dashboard' }
  );

  breadcrumbs = computed(() => {
    const url = this.routeUrl();
    const routeMap: Record<string, string> = {
      '/dashboard': 'Overview',
      '/workloads': 'Workloads',
      '/builds': 'Build Pipeline',
      '/cluster': 'Cluster Topology',
      '/raft': 'Raft Engine',
      '/replication': 'Replication',
      '/failovers': 'Failover Events',
      '/settings': 'Settings',
      '/onboarding': 'Onboarding',
    };
    return ['RaftWeave', routeMap[url] || 'Dashboard'];
  });
}
