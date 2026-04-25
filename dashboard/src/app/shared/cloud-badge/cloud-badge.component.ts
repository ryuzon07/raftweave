import { Component, input } from '@angular/core';
import { CloudProvider } from '../../core/models/types';

@Component({
  selector: 'rw-cloud-badge',
  standalone: true,
  template: `
    <span class="badge">
      <span class="badge__icon material-symbols-outlined" [style.color]="cloudColor()">cloud</span>
      <span class="badge__label">{{ region() }}</span>
    </span>
  `,
  styles: [`
    .badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      background: rgba(81, 82, 162, 0.4);
      border: 1px solid rgba(91, 106, 189, 0.2);
      border-radius: 6px;
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      white-space: nowrap;
    }
    .badge__icon {
      font-size: 1rem;
    }
  `]
})
export class CloudBadgeComponent {
  cloud = input<CloudProvider>('AWS');
  region = input<string>('');

  cloudColor(): string {
    switch (this.cloud()) {
      case 'AWS': return '#F59E0B'; // Amber
      case 'Azure': return '#3B82F6'; // Blue
      case 'GCP': return '#EF4444'; // Red
      default: return 'var(--rw-mist)';
    }
  }
}
