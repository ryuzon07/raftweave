import { Component, input, ElementRef, viewChild, AfterViewInit, OnDestroy, effect } from '@angular/core';
import { BuildLogLine } from '../../core/models/types';

@Component({
  selector: 'rw-log-stream',
  standalone: true,
  template: `
    <div class="log-wrapper">
      <div class="log-header">
        <span class="log-header__title">{{ title() }}</span>
        <button class="log-header__toggle" (click)="scrollLocked = !scrollLocked">
          <span class="material-symbols-outlined" style="font-size: 14px; vertical-align: text-bottom;">
            {{ scrollLocked ? 'lock' : 'lock_open' }}
          </span>
          {{ scrollLocked ? 'Locked' : 'Unlocked' }}
        </button>
      </div>
      <div class="log-content rw-terminal" #logContainer>
        @for (line of lines(); track $index) {
          <div class="log-line">
            <span class="timestamp">{{ formatTime(line.timestamp) }}</span>
            <span [class]="line.level">{{ line.line }}</span>
          </div>
        }
        @if (lines().length === 0) {
          <div class="log-empty">Waiting for log data...</div>
        }
      </div>
    </div>
  `,
  styles: [`
    .log-wrapper {
      display: flex;
      flex-direction: column;
      border-radius: var(--radius-md);
      overflow: hidden;
      border: 1px solid rgba(91, 106, 189, 0.15);
    }
    .log-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 16px;
      background: rgba(10, 10, 18, 0.8);
      border-bottom: 1px solid rgba(91, 106, 189, 0.1);
    }
    .log-header__title {
      font-family: var(--font-heading);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .log-header__toggle {
      background: transparent;
      border: 1px solid rgba(91, 106, 189, 0.3);
      color: var(--rw-mist);
      padding: 2px 8px;
      border-radius: 4px;
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      cursor: pointer;
    }
    .log-content {
      height: 300px;
      overflow-y: auto;
      padding: 12px 16px;
    }
    .log-line {
      display: flex;
      gap: 4px;
      padding: 1px 0;
    }
    .log-empty {
      color: var(--rw-mid);
      font-style: italic;
      padding: 20px 0;
      text-align: center;
    }
  `]
})
export class LogStreamComponent implements AfterViewInit, OnDestroy {
  title = input<string>('Build Logs');
  lines = input<BuildLogLine[]>([]);

  logContainer = viewChild<ElementRef>('logContainer');
  scrollLocked = true;

  private scrollEffect = effect(() => {
    const _ = this.lines(); // Track signal
    if (this.scrollLocked) {
      setTimeout(() => this.scrollToBottom(), 10);
    }
  });

  ngAfterViewInit() {
    this.scrollToBottom();
  }

  ngOnDestroy() {}

  private scrollToBottom(): void {
    const el = this.logContainer()?.nativeElement;
    if (el) el.scrollTop = el.scrollHeight;
  }

  formatTime(ts: string): string {
    try {
      return new Date(ts).toLocaleTimeString('en-US', { hour12: false });
    } catch {
      return '--:--:--';
    }
  }
}
