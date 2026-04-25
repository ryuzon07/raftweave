import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from './layout/sidebar/sidebar.component';
import { TopbarComponent } from './layout/topbar/topbar.component';
import { ToastContainerComponent } from './shared/toast/toast-container.component';
import { AnimationService } from './core/services/animation.service';
import { MockDataService } from './core/services/mock-data.service';

import { AuthService } from './core/services/auth.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, SidebarComponent, TopbarComponent, ToastContainerComponent],
  template: `
    <!-- 8-bit Cloud Particle Background -->
    <div class="cloud-scene" aria-hidden="true">
      <div class="cloud-scene__gradient"></div>
      <div class="cloud-layer cloud-layer--back"></div>
      <div class="cloud-layer cloud-layer--mid"></div>
      <div class="cloud-layer cloud-layer--front"></div>
    </div>

    @if (auth.isAuthenticated()) {
      <rw-sidebar />
      <rw-topbar />
      <main class="main-content">
        <router-outlet />
      </main>
    } @else {
      <main class="login-content">
        <router-outlet />
      </main>
    }

    <rw-toast-container />
  `,
  styles: [`
    /* ━━━ 8-Bit Cloud Particle Scene ━━━ */
    .cloud-scene {
      position: fixed;
      inset: 0;
      z-index: 0;
      pointer-events: none;
      overflow: hidden;
    }

    /* Subtle radial "light source" gradient */
    .cloud-scene__gradient {
      position: absolute;
      inset: 0;
      background: radial-gradient(
        ellipse 70% 50% at 55% 40%,
        rgba(119, 126, 240, 0.06) 0%,
        transparent 70%
      );
    }

    /* Cloud layers — 8-bit pixelated silhouettes via CSS box-shadow "sprites" */
    .cloud-layer {
      position: absolute;
      top: 0;
      left: 0;
      width: 200%;
      height: 100%;
      image-rendering: pixelated;
      image-rendering: crisp-edges;
    }

    /* Back layer — large, slow, very dim */
    .cloud-layer--back {
      opacity: 0.04;
      animation: rw-cloud-drift-slow 120s linear infinite;
      background-image:
        radial-gradient(circle 40px at 10% 20%, rgba(91,106,189,1) 0%, transparent 100%),
        radial-gradient(circle 60px at 30% 70%, rgba(91,106,189,1) 0%, transparent 100%),
        radial-gradient(circle 50px at 55% 15%, rgba(91,106,189,1) 0%, transparent 100%),
        radial-gradient(circle 45px at 75% 55%, rgba(91,106,189,1) 0%, transparent 100%),
        radial-gradient(circle 55px at 90% 35%, rgba(91,106,189,1) 0%, transparent 100%);
    }

    /* Mid layer — medium, moderate speed */
    .cloud-layer--mid {
      opacity: 0.05;
      animation: rw-cloud-drift-mid 80s linear infinite;
      background-image:
        radial-gradient(circle 30px at 8% 45%, rgba(119,126,240,1) 0%, transparent 100%),
        radial-gradient(circle 40px at 25% 25%, rgba(119,126,240,1) 0%, transparent 100%),
        radial-gradient(circle 35px at 48% 60%, rgba(119,126,240,1) 0%, transparent 100%),
        radial-gradient(circle 28px at 65% 30%, rgba(119,126,240,1) 0%, transparent 100%),
        radial-gradient(circle 38px at 85% 65%, rgba(119,126,240,1) 0%, transparent 100%);
    }

    /* Front layer — small, faster, brightest (but still ghosted) */
    .cloud-layer--front {
      opacity: 0.06;
      animation: rw-cloud-drift-fast 50s linear infinite;
      background-image:
        radial-gradient(circle 18px at 5% 55%, rgba(227,233,255,1) 0%, transparent 100%),
        radial-gradient(circle 22px at 20% 35%, rgba(227,233,255,1) 0%, transparent 100%),
        radial-gradient(circle 16px at 40% 75%, rgba(227,233,255,1) 0%, transparent 100%),
        radial-gradient(circle 20px at 60% 20%, rgba(227,233,255,1) 0%, transparent 100%),
        radial-gradient(circle 24px at 80% 50%, rgba(227,233,255,1) 0%, transparent 100%),
        radial-gradient(circle 14px at 92% 80%, rgba(227,233,255,1) 0%, transparent 100%);
    }

    /* ━━━ Main content area ━━━ */
    .main-content {
      position: relative;
      z-index: 1;
      margin-left: var(--sidebar-width);
      margin-top: var(--topbar-height);
      min-height: calc(100vh - var(--topbar-height));
      padding: 24px;
      transition: margin-left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    }

    @media (max-width: 768px) {
      .main-content {
        margin-left: 0;
        padding: 16px;
        padding-bottom: 80px;
      }
    }

    .login-content {
      position: relative;
      z-index: 10;
      width: 100vw;
      min-height: 100vh;
    }
  `]
})
export class App implements OnInit, OnDestroy {
  auth = inject(AuthService);
  private animService = inject(AnimationService);
  private mockData = inject(MockDataService);

  ngOnInit(): void {
    this.animService.initLenis();
    this.mockData.startStreaming();
  }

  ngOnDestroy(): void {
    this.animService.destroyLenis();
    this.mockData.stopStreaming();
  }
}
