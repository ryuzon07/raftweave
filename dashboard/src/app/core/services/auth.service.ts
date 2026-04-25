import { Injectable, signal, WritableSignal, Signal } from '@angular/core';

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

  readonly user: Signal<User | null> = this._user.asReadonly();
  readonly isAuthenticated: Signal<boolean> = this._isAuthenticated.asReadonly();


  login(provider: 'github' | 'google'): void {
    // Simulated OAuth PKCE flow
    this._user.set({
      id: 'user-001',
      name: 'Prithviraj',
      email: 'prithviraj@raftweave.io',
      avatar: '',
      role: 'admin'
    });
    this._isAuthenticated.set(true);
  }

  logout(): void {
    this._user.set(null);
    this._isAuthenticated.set(false);
  }

  canPerformAction(action: 'deploy' | 'scale' | 'delete' | 'settings'): boolean {
    const user = this._user();
    if (!user) return false;
    if (user.role === 'admin') return true;
    if (user.role === 'operator') return action !== 'settings';
    return false; // viewer
  }
}
