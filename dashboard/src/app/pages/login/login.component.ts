import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'rw-login',
  standalone: true,
  template: `
    <div class="login-wrapper">
      <!-- Immersive background -->
      <div class="login-bg">
        <div class="login-bg__gradient"></div>
        <div class="login-bg__grid"></div>
        
        <!-- Floating 8-bit Clouds -->
        <div class="login-clouds">
          <div class="login-cloud cloud-1"></div>
          <div class="login-cloud cloud-2"></div>
          <div class="login-cloud cloud-3"></div>
          <div class="login-cloud cloud-4"></div>
        </div>
      </div>

      <div class="login-left">
        <div class="login-left__content">
          <div class="login-logo-wrap">
            <img src="/assets/raftweave-logo-transparent.png" alt="RaftWeave Logo" class="login-logo-img">
          </div>
          <p class="login-tagline">Enterprise-grade multi-cloud deployment<br>powered by Raft consensus.</p>
        </div>
      </div>

      <div class="login-right">
        <div class="login-form-panel">
          <div class="login-form__header">
            <h2 class="login-form__title">Welcome back</h2>
            <p class="login-form__subtitle">Authenticate to access your deployment console</p>
          </div>

          <div class="login-form__actions">
            <button class="login-btn github" (click)="loginWith('github')">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
              </svg>
              Continue with GitHub
            </button>

            <button class="login-btn google" (click)="loginWith('google')">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
              </svg>
              Continue with Google
            </button>
          </div>

          <div class="login-form__divider">
            <span>Secure connection established</span>
          </div>

          <p class="login-form__footer">
            Protected by RaftWeave Identity • <span>v2.4.0</span>
          </p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    :host { 
      display: block; 
      position: fixed;
      inset: 0;
      z-index: 9999;
    }
    
    .login-wrapper {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 100px;
      width: 100vw;
      height: 100vh;
      background: var(--rw-deep);
      position: relative;
      padding: 40px;
    }

    /* Immersive Background */
    .login-bg {
      position: absolute;
      inset: 0;
      overflow: hidden;
      pointer-events: none;
      z-index: 0;
    }

    .login-bg__gradient {
      position: absolute;
      inset: 0;
      background: radial-gradient(circle at 50% 50%, rgba(119, 126, 240, 0.12), transparent 70%);
    }

    .login-bg__grid {
      position: absolute;
      inset: -50% -50% 100% -50%;
      background-image:
        linear-gradient(rgba(119, 126, 240, 0.12) 1px, transparent 1px),
        linear-gradient(90deg, rgba(119, 126, 240, 0.12) 1px, transparent 1px);
      background-size: 80px 80px;
      transform: perspective(1000px) rotateX(60deg) translateY(-100px) translateZ(-200px);
      transform-origin: top center;
      animation: grid-move 20s linear infinite;
    }

    @keyframes grid-move {
      0% { background-position: 0 0; }
      100% { background-position: 0 80px; }
    }

    .login-left {
      flex: 0 1 500px;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      position: relative;
      z-index: 1;
      text-align: center;
    }

    .login-left__content {
      display: flex;
      flex-direction: column;
      align-items: center;
      animation: rw-fade-in 1s cubic-bezier(0.16, 1, 0.3, 1);
    }

    .login-clouds {
      position: absolute;
      inset: 0;
      overflow: hidden;
      pointer-events: none;
      z-index: 1;
    }

    .login-cloud {
      position: absolute;
      background-image: url('/assets/login-cloud-v2.png');
      background-size: contain;
      background-repeat: no-repeat;
      image-rendering: pixelated;
      opacity: 0.2;
      mix-blend-mode: screen;
    }

    .cloud-1 {
      width: 800px;
      height: 600px;
      top: 5%;
      left: -10%;
      animation: rw-drift-slow 180s linear infinite;
    }

    .cloud-2 {
      width: 600px;
      height: 450px;
      top: 40%;
      left: 20%;
      opacity: 0.12;
      animation: rw-drift-mid 120s linear infinite reverse;
    }

    .cloud-3 {
      width: 1000px;
      height: 750px;
      top: 60%;
      left: -20%;
      animation: rw-drift-slow 220s linear infinite;
      animation-delay: -50s;
    }

    .cloud-4 {
      width: 500px;
      height: 380px;
      top: 15%;
      right: -5%;
      opacity: 0.15;
      animation: rw-drift-fast 90s linear infinite;
      animation-delay: -20s;
    }

    @keyframes rw-drift-slow {
      0% { transform: translateX(-100%) translateY(0); }
      100% { transform: translateX(200%) translateY(10px); }
    }

    @keyframes rw-drift-mid {
      0% { transform: translateX(-100%) translateY(0); }
      100% { transform: translateX(200%) translateY(-20px); }
    }

    @keyframes rw-drift-fast {
      0% { transform: translateX(-100%) translateY(0); }
      100% { transform: translateX(200%) translateY(30px); }
    }

    .login-logo-wrap {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 32px;
      animation: rw-float 6s ease-in-out infinite;
    }

    .login-logo-img {
      width: 380px;
      height: auto;
      object-fit: contain;
      filter: drop-shadow(0 0 32px rgba(119, 126, 240, 0.3));
    }

    .login-tagline {
      font-family: var(--font-body);
      font-size: 1.25rem;
      color: var(--rw-mist);
      line-height: 1.6;
      max-width: 400px;
      opacity: 0.9;
    }

    .login-right {
      flex: 0 0 420px;
      display: flex;
      align-items: center;
      justify-content: center;
      position: relative;
      z-index: 1;
    }

    .login-form-panel {
      width: 100%;
      background: rgba(11, 14, 20, 0.7);
      backdrop-filter: blur(24px);
      -webkit-backdrop-filter: blur(24px);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: var(--radius-xl);
      padding: 56px 48px;
      box-shadow: 
        0 32px 64px rgba(0, 0, 0, 0.4),
        inset 0 1px 0 rgba(255, 255, 255, 0.1);
      animation: rw-slide-in-right 0.8s cubic-bezier(0.16, 1, 0.3, 1);
    }

    .login-form__header {
      margin-bottom: 48px;
      text-align: center;
    }

    .login-form__title {
      font-family: var(--font-heading);
      font-size: 1.75rem;
      font-weight: 700;
      color: var(--rw-white);
      margin-bottom: 12px;
      letter-spacing: -0.01em;
    }

    .login-form__subtitle {
      font-family: var(--font-body);
      font-size: var(--text-body);
      color: var(--rw-mid);
      line-height: 1.5;
    }

    .login-form__actions {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }

    .login-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;
      width: 100%;
      padding: 16px 24px;
      border-radius: var(--radius-md);
      font-family: var(--font-body);
      font-size: 1rem;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
      background: rgba(255, 255, 255, 0.05);
      border: 1px solid rgba(255, 255, 255, 0.12);
      color: var(--rw-white);
      position: relative;
      overflow: hidden;
    }

    .login-btn::before {
      content: '';
      position: absolute;
      inset: 0;
      background: linear-gradient(90deg, transparent, rgba(255,255,255,0.08), transparent);
      transform: translateX(-100%);
      transition: transform 0.6s ease;
    }

    .login-btn:hover::before {
      transform: translateX(100%);
    }

    .login-btn:hover {
      background: rgba(255, 255, 255, 0.12);
      border-color: rgba(255, 255, 255, 0.3);
      transform: translateY(-2px);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
    }

    .login-btn svg {
      opacity: 0.9;
      transition: transform 0.3s ease;
    }

    .login-btn:hover svg {
      transform: scale(1.1);
    }

    .login-form__divider {
      margin: 32px 0;
      text-align: center;
      position: relative;
    }

    .login-form__divider::before {
      content: '';
      position: absolute;
      top: 50%;
      left: 0;
      right: 0;
      height: 1px;
      background: linear-gradient(90deg, transparent, rgba(110, 231, 183, 0.2), transparent);
    }

    .login-form__divider span {
      background: rgba(40, 42, 85, 1);
      padding: 0 12px;
      color: var(--healthy);
      font-family: var(--font-mono);
      font-size: 0.7rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      position: relative;
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }

    .login-form__divider span::before {
      content: '';
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: var(--healthy);
      box-shadow: 0 0 8px var(--healthy);
      animation: rw-pulse-dot 2s infinite;
    }

    .login-form__footer {
      font-family: var(--font-mono);
      font-size: 0.7rem;
      color: var(--rw-mid);
      text-align: center;
    }

    .login-form__footer span {
      color: var(--rw-mist);
    }

    @media (max-width: 900px) {
      .login-left { display: none; }
      .login-right { flex: 1; padding: 24px; }
      .login-form-panel { padding: 40px 24px; }
      .login-bg__grid { transform: perspective(1000px) rotateX(60deg) translateY(0) translateZ(-100px); }
    }
  `]
})
export class LoginComponent {
  private auth = inject(AuthService);
  private router = inject(Router);

  loginWith(provider: 'github' | 'google'): void {
    this.auth.login(provider);
    this.router.navigate(['/dashboard']);
  }
}
