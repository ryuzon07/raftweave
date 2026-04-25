import { Component, signal, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { AnimationService } from '../../core/services/animation.service';

interface NavItem {
  icon: string;
  label: string;
  route: string;
}

@Component({
  selector: 'rw-sidebar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  template: `
    <aside class="sidebar" [class.collapsed]="collapsed()">
      <!-- Logo -->
      <div class="sidebar__logo" (click)="toggleSidebar()">
        @if (!collapsed()) {
          <img src="/assets/raftweave-logo-transparent.png" alt="RaftWeave" class="sidebar__logo-img">
        } @else {
          <span class="sidebar__logo-icon">◆</span>
        }
      </div>

      <!-- Workspace -->
      @if (!collapsed()) {
        <div class="sidebar__workspace">
          <div class="sidebar__ws-dot-wrap">
            <div class="sidebar__ws-dot"></div>
            <div class="sidebar__ws-ping"></div>
          </div>
          <span class="sidebar__ws-name">Production</span>
        </div>
      }

      <div class="sidebar__divider"></div>

      <!-- Navigation -->
      <nav class="sidebar__nav">
        @for (item of navItems; track item.route) {
          <a class="sidebar__item"
             [routerLink]="item.route"
             routerLinkActive="active"
             [routerLinkActiveOptions]="{ exact: item.route === '/dashboard' }"
             (mouseenter)="onItemHover($event, true)"
             (mouseleave)="onItemHover($event, false)">
            <span class="sidebar__item-icon material-symbols-outlined">{{ item.icon }}</span>
            @if (!collapsed()) {
              <span class="sidebar__item-label">{{ item.label }}</span>
            }
          </a>
        }
      </nav>

      <!-- Collapse Toggle -->
      <div class="sidebar__footer">
        <button class="sidebar__toggle" (click)="toggleSidebar()">
          <span class="sidebar__toggle-icon">{{ collapsed() ? '→' : '←' }}</span>
          @if (!collapsed()) {
            <span class="sidebar__toggle-label">Collapse</span>
          }
        </button>
      </div>
    </aside>

    <!-- Mobile Bottom Nav -->
    <nav class="mobile-nav">
      @for (item of navItems.slice(0, 5); track item.route) {
        <a class="mobile-nav__item"
           [routerLink]="item.route"
           routerLinkActive="active">
          <span class="mobile-nav__icon material-symbols-outlined">{{ item.icon }}</span>
          <span class="mobile-nav__label">{{ item.label }}</span>
        </a>
      }
    </nav>
  `,
  styles: [`
    .sidebar {
      position: fixed;
      top: 0;
      left: 0;
      width: var(--sidebar-width);
      height: 100vh;
      background: rgba(56, 58, 110, 0.65);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border-right: 1px solid rgba(255, 255, 255, 0.06);
      display: flex;
      flex-direction: column;
      padding: 12px;
      z-index: 100;
      transition: width 0.3s cubic-bezier(0.16, 1, 0.3, 1);
      overflow: hidden;
    }
    .sidebar.collapsed {
      width: var(--sidebar-collapsed);
    }
    .sidebar__logo {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 8px 8px;
      cursor: pointer;
      margin-bottom: 8px;
    }
    .sidebar__logo-icon {
      font-size: 1.3rem;
      color: var(--rw-accent);
      flex-shrink: 0;
      text-shadow: 0 0 12px rgba(119, 126, 240, 0.5);
      margin: 0 auto;
    }
    .sidebar__logo-img {
      width: 160px;
      height: auto;
      object-fit: contain;
      filter: drop-shadow(0 0 12px rgba(255, 255, 255, 0.15));
      margin-left: -4px;
    }
    .sidebar__workspace {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 6px 8px;
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: var(--radius-sm);
      margin-bottom: 8px;
    }
    .sidebar__ws-dot-wrap {
      position: relative;
      width: 8px;
      height: 8px;
      flex-shrink: 0;
    }
    .sidebar__ws-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--healthy);
      box-shadow: 0 0 6px rgba(110, 231, 183, 0.5);
      animation: rw-breathe 3s ease-in-out infinite;
      position: relative;
      z-index: 1;
    }
    .sidebar__ws-ping {
      position: absolute;
      inset: 0;
      border-radius: 50%;
      background: var(--healthy);
      animation: rw-ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
    }
    .sidebar__ws-name {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      white-space: nowrap;
    }
    .sidebar__divider {
      height: 1px;
      background: rgba(255, 255, 255, 0.06);
      margin: 4px 0 8px;
    }
    .sidebar__nav {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }
    .sidebar__item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 8px 10px;
      color: var(--rw-mist);
      border-radius: var(--radius-sm);
      font-family: var(--font-body);
      font-size: var(--text-body);
      font-weight: 400;
      text-decoration: none;
      transition: all 0.3s ease-out;
      position: relative;
      white-space: nowrap;
    }
    .sidebar__item:hover {
      color: var(--rw-white);
      background: rgba(255, 255, 255, 0.06);
    }

    /* ━━━ Active Item — Neon Glow ━━━ */
    .sidebar__item.active {
      color: var(--rw-white);
      background: linear-gradient(90deg, rgba(119, 126, 240, 0.15) 0%, rgba(119, 126, 240, 0.04) 100%);
      box-shadow: 0 0 16px rgba(119, 126, 240, 0.08);
    }
    .sidebar__item.active::before {
      content: '';
      position: absolute;
      left: -12px;
      top: 4px;
      bottom: 4px;
      width: 3px;
      background: var(--rw-accent);
      border-radius: 0 2px 2px 0;
      box-shadow: 0 0 8px rgba(119, 126, 240, 0.6), 0 0 16px rgba(119, 126, 240, 0.3);
    }
    .sidebar__item.active .sidebar__item-icon {
      color: var(--rw-accent);
      text-shadow: 0 0 10px rgba(119, 126, 240, 0.5);
    }

    .sidebar__item-icon {
      font-size: 1.1rem;
      flex-shrink: 0;
      width: 24px;
      text-align: center;
      transition: color 0.3s ease-out, text-shadow 0.3s ease-out;
    }
    .sidebar__item-label {
      white-space: nowrap;
    }
    .sidebar__footer {
      padding-top: 8px;
      border-top: 1px solid rgba(255, 255, 255, 0.06);
    }
    .sidebar__toggle {
      display: flex;
      align-items: center;
      gap: 8px;
      width: 100%;
      padding: 8px 10px;
      background: transparent;
      border: none;
      color: var(--rw-mid);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      cursor: pointer;
      border-radius: var(--radius-sm);
      transition: all 0.3s ease-out;
    }
    .sidebar__toggle:hover {
      color: var(--rw-mist);
      background: rgba(255, 255, 255, 0.04);
    }
    .sidebar__toggle-icon { font-size: 0.9rem; }

    /* Mobile bottom nav */
    .mobile-nav {
      display: none;
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      background: rgba(56, 58, 110, 0.85);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border-top: 1px solid rgba(255, 255, 255, 0.06);
      padding: 6px 0 env(safe-area-inset-bottom, 6px);
      z-index: 100;
    }
    @media (max-width: 768px) {
      .sidebar { display: none; }
      .mobile-nav {
        display: flex;
        justify-content: space-around;
      }
    }
    .mobile-nav__item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 2px;
      padding: 4px 8px;
      color: var(--rw-mid);
      text-decoration: none;
      font-size: var(--text-xs);
      transition: color 0.3s ease-out;
    }
    .mobile-nav__item.active {
      color: var(--rw-accent);
    }
    .mobile-nav__icon { font-size: 1.2rem; }
    .mobile-nav__label {
      font-family: var(--font-body);
      font-size: 0.6rem;
    }
  `]
})
export class SidebarComponent {
  private animService = inject(AnimationService);

  collapsed = signal(false);

  toggleSidebar() {
    const isCollapsed = !this.collapsed();
    this.collapsed.set(isCollapsed);
    document.documentElement.style.setProperty('--sidebar-width', isCollapsed ? '64px' : '240px');
  }

  navItems: NavItem[] = [
    { icon: 'home', label: 'Overview', route: '/dashboard' },
    { icon: 'inventory_2', label: 'Workloads', route: '/workloads' },
    { icon: 'build', label: 'Build Pipeline', route: '/builds' },
    { icon: 'language', label: 'Cluster Topology', route: '/cluster' },
    { icon: 'settings_applications', label: 'Raft Engine', route: '/raft' },
    { icon: 'bar_chart', label: 'Replication', route: '/replication' },
    { icon: 'electric_bolt', label: 'Failover Events', route: '/failovers' },
    { icon: 'settings', label: 'Settings', route: '/settings' },
  ];

  onItemHover(event: MouseEvent, enter: boolean): void {
    const icon = (event.currentTarget as HTMLElement).querySelector('.sidebar__item-icon');
    if (icon) {
      this.animService.iconHover(icon, enter);
    }
  }
}
