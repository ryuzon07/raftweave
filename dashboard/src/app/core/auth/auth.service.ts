import { Injectable, signal, computed } from '@angular/core';
import { Router } from '@angular/router';

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  avatarUrl: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly currentUser = signal<AuthUser | null>(null);

  readonly user = this.currentUser.asReadonly();
  readonly isAuthenticated = computed(() => this.currentUser() !== null);

  constructor(private readonly router: Router) {}

  async login(provider: 'github' | 'google'): Promise<void> {
    // TODO: Implement OAuth login flow via Connect-RPC
    throw new Error(`Login with ${provider} not implemented`);
  }

  async logout(): Promise<void> {
    this.currentUser.set(null);
    await this.router.navigate(['/']);
  }

  setUser(user: AuthUser): void {
    this.currentUser.set(user);
  }
}
