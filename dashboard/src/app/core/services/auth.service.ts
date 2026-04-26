import { Injectable, signal, WritableSignal, Signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';

export interface User {
  id: string;
  name: string;
  email: string;
  avatar: string;
  role: 'admin' | 'operator' | 'viewer';
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly _user: WritableSignal<User | null> = signal(null);
  private readonly _isAuthenticated: WritableSignal<boolean> = signal(false);
  private readonly _isInitialized: WritableSignal<boolean> = signal(false);
  private readonly http = inject(HttpClient);

  readonly user: Signal<User | null> = this._user.asReadonly();
  readonly isAuthenticated: Signal<boolean> = this._isAuthenticated.asReadonly();
  readonly isInitialized: Signal<boolean> = this._isInitialized.asReadonly();

  constructor() {
    this.checkAuth();
  }

  checkAuth(): void {
    this.http.post<any>(`${environment.apiBaseUrl}/auth.v1.AuthService/GetMe`, {}).subscribe({
      next: (response) => {
        const u = response.user || response;
        this._user.set({
          id: u.id || 'user-001',
          name: u.name || u.email?.split('@')[0] || 'Unknown',
          email: u.email || '',
          avatar: u.avatarUrl || '',
          role: u.role?.toLowerCase() || 'admin'
        });
        this._isAuthenticated.set(true);
        this._isInitialized.set(true);
      },
      error: () => {
        this._user.set(null);
        this._isAuthenticated.set(false);
        this._isInitialized.set(true);
      }
    });
  }

  login(provider: 'github' | 'google'): void {
    window.location.href = `${environment.apiBaseUrl}/auth/${provider}/login`;
  }

  logout(): void {
    this._user.set(null);
    this._isAuthenticated.set(false);
    // Call backend logout or clear cookies if needed
    window.location.href = '/';
  }

  canPerformAction(action: 'deploy' | 'scale' | 'delete' | 'settings'): boolean {
    const user = this._user();
    if (!user) return false;
    if (user.role === 'admin') return true;
    if (user.role === 'operator') return action !== 'settings';
    return false; // viewer
  }
}
