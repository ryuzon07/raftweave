import { Injectable, signal, WritableSignal, computed, Signal } from '@angular/core';

export interface ToastMessage {
  id: string;
  title: string;
  body: string;
  type: 'info' | 'success' | 'warning' | 'error' | 'election';
  action?: { label: string; callback: () => void };
  autoDismissMs?: number;
}

@Injectable({ providedIn: 'root' })
export class ToastService {
  private readonly _toasts: WritableSignal<ToastMessage[]> = signal([]);
  readonly toasts: Signal<ToastMessage[]> = this._toasts.asReadonly();

  show(msg: Omit<ToastMessage, 'id'>): string {
    const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    const toast: ToastMessage = { ...msg, id };
    this._toasts.update(t => [toast, ...t]);

    const dismissMs = msg.autoDismissMs ?? 5000;
    if (dismissMs > 0) {
      setTimeout(() => this.dismiss(id), dismissMs);
    }
    return id;
  }

  dismiss(id: string): void {
    this._toasts.update(t => t.filter(x => x.id !== id));
  }

  /** Convenience methods */
  election(title: string, body: string): void {
    this.show({ title, body, type: 'election', autoDismissMs: 8000 });
  }

  buildComplete(workload: string, digest: string): void {
    this.show({
      title: `Build Complete — ${workload}`,
      body: `Image: ${digest}`,
      type: 'success',
      action: { label: 'View Logs', callback: () => {} }
    });
  }

  error(title: string, body: string): void {
    this.show({ title, body, type: 'error', autoDismissMs: 8000 });
  }
}
