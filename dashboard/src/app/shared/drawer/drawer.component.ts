import { Component, input, output } from '@angular/core';

@Component({
  selector: 'rw-drawer',
  standalone: true,
  template: `
    @if (open()) {
      <div class="drawer-backdrop" (click)="closed.emit()"></div>
      <div class="drawer" [class.open]="open()">
        <div class="drawer__header">
          <h3 class="drawer__title">{{ title() }}</h3>
          <button class="drawer__close" (click)="closed.emit()">×</button>
        </div>
        <div class="drawer__content">
          <ng-content />
        </div>
      </div>
    }
  `,
  styles: [`
    .drawer-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.5);
      backdrop-filter: blur(4px);
      z-index: 1000;
      animation: rw-fade-in 0.2s ease;
    }
    .drawer {
      position: fixed;
      top: 0;
      right: 0;
      width: min(560px, 90vw);
      height: 100vh;
      background: var(--surface-1);
      border-left: 1px solid rgba(91, 106, 189, 0.2);
      z-index: 1001;
      display: flex;
      flex-direction: column;
      animation: rw-slide-in-right 0.3s cubic-bezier(0.16, 1, 0.3, 1);
      box-shadow: -8px 0 32px rgba(0, 0, 0, 0.3);
    }
    .drawer__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 24px;
      border-bottom: 1px solid rgba(91, 106, 189, 0.15);
    }
    .drawer__title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      font-weight: 600;
      color: var(--rw-white);
    }
    .drawer__close {
      background: transparent;
      border: none;
      color: var(--rw-mist);
      font-size: 1.5rem;
      cursor: pointer;
      padding: 4px 8px;
      line-height: 1;
      border-radius: var(--radius-sm);
      transition: all var(--transition-fast);
    }
    .drawer__close:hover {
      color: var(--rw-white);
      background: rgba(91, 106, 189, 0.2);
    }
    .drawer__content {
      flex: 1;
      overflow-y: auto;
      padding: 24px;
    }
  `]
})
export class DrawerComponent {
  open = input<boolean>(false);
  title = input<string>('');
  closed = output<void>();
}
