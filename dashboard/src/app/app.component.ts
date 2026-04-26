import { Component, inject, effect, signal } from '@angular/core';
import { RouterOutlet, Router, NavigationEnd } from '@angular/router';
import { SidebarComponent } from './shared/components/sidebar/sidebar.component';
import { HeaderComponent } from './shared/components/header/header.component';
import { AuthService } from './core/services/auth.service';
import { CommonModule } from '@angular/common';
import { filter } from 'rxjs/operators';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, SidebarComponent, HeaderComponent, CommonModule],
  template: `
    @if (isDashboardVisible()) {
      <div class="flex h-screen bg-gray-900">
        <app-sidebar />
        <div class="flex flex-col flex-1 overflow-hidden">
          <app-header />
          <main class="flex-1 overflow-y-auto">
            <router-outlet />
          </main>
        </div>
      </div>
    } @else {
      <div class="h-screen bg-gray-950 flex items-center justify-center">
        <router-outlet />
      </div>
    }
  `,
})
export class AppComponent {
  auth = inject(AuthService);
  router = inject(Router);
  // Initialize with current URL immediately
  currentUrl = signal(window.location.pathname);
  title = 'RaftWeave';

  isDashboardVisible(): boolean {
    const url = this.currentUrl();
    // Hide shell if on login or if not authenticated
    const isLoginPage = url.includes('/login') || url === '/';
    return this.auth.isAuthenticated() && !isLoginPage;
  }

  constructor() {
    // Track URL changes reactively
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe((event: any) => {
      this.currentUrl.set(event.urlAfterRedirects);
    });

    effect(() => {
      console.log('App State -> Auth:', this.auth.isAuthenticated(), 'URL:', this.currentUrl());
    });
  }
}
