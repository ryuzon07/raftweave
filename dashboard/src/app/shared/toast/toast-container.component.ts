import { Component, inject } from '@angular/core';
import { ToastService, ToastMessage } from '../../core/services/toast.service';

@Component({
  selector: 'rw-toast-container',
  standalone: true,
  template: `
    <div class="toast-container">
      @for (toast of toastService.toasts(); track toast.id) {
        <div class="toast" [class]="'toast--' + toast.type">
          <div class="toast__icon material-symbols-outlined">
            @switch (toast.type) {
              @case ('election') { electric_bolt }
              @case ('success') { check_circle }
              @case ('warning') { warning }
              @case ('error') { error }
              @default { info }
            }
          </div>
          <div class="toast__content">
            <div class="toast__title">{{ toast.title }}</div>
            <div class="toast__body">{{ toast.body }}</div>
          </div>
          @if (toast.action) {
            <button class="toast__action" (click)="toast.action!.callback()">
              {{ toast.action!.label }}
            </button>
          }
          <button class="toast__close" (click)="toastService.dismiss(toast.id)">×</button>
        </div>
      }
    </div>
  `,
  styles: [`
    .toast-container {
      position: fixed;
      top: 60px;
      right: 16px;
      z-index: 9999;
      display: flex;
      flex-direction: column;
      gap: 8px;
      max-width: 400px;
    }
    .toast {
      display: flex;
      align-items: flex-start;
      gap: 10px;
      padding: 12px 16px;
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.3);
      border-radius: var(--radius-md);
      backdrop-filter: blur(16px);
      animation: rw-fade-in 0.25s ease;
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
    }
    .toast--election { border-left: 3px solid var(--rw-accent); }
    .toast--success { border-left: 3px solid var(--healthy); }
    .toast--warning { border-left: 3px solid var(--degraded); }
    .toast--error { border-left: 3px solid var(--failed); }
    .toast__icon { font-size: 1.1rem; padding-top: 1px; }
    .toast__content { flex: 1; min-width: 0; }
    .toast__title {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      font-weight: 600;
      color: var(--rw-white);
    }
    .toast__body {
      font-size: var(--text-xs);
      color: var(--rw-mist);
      margin-top: 2px;
    }
    .toast__action {
      background: transparent;
      border: 1px solid var(--rw-accent);
      color: var(--rw-accent);
      padding: 4px 10px;
      border-radius: var(--radius-sm);
      font-size: var(--text-xs);
      font-family: var(--font-body);
      cursor: pointer;
      white-space: nowrap;
      transition: all var(--transition-fast);
    }
    .toast__action:hover {
      background: var(--rw-accent);
      color: var(--rw-white);
    }
    .toast__close {
      background: transparent;
      border: none;
      color: var(--rw-mid);
      font-size: 1.1rem;
      cursor: pointer;
      padding: 0 4px;
      line-height: 1;
    }
    .toast__close:hover { color: var(--rw-white); }
  `]
})
export class ToastContainerComponent {
  toastService = inject(ToastService);
}
