import { Component } from '@angular/core';
import { StatusBadgeComponent } from '../status-badge/status-badge.component';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [StatusBadgeComponent],
  template: `
    <header class="bg-white shadow px-6 py-4 flex items-center justify-between">
      <h1 class="text-lg font-semibold text-gray-800">Control Plane</h1>
      <div class="flex items-center space-x-4">
        <app-status-badge [status]="'healthy'" />
        <span class="text-sm text-gray-600">Cluster Online</span>
      </div>
    </header>
  `,
})
export class HeaderComponent {}
