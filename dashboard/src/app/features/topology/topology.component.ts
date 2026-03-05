import { Component } from '@angular/core';

@Component({
  selector: 'app-topology',
  standalone: true,
  template: `
    <div>
      <h2 class="text-2xl font-bold mb-6">Multi-Cloud Topology</h2>
      <div class="grid grid-cols-3 gap-6">
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="font-semibold text-lg mb-2">AWS</h3>
          <p class="text-gray-500">ap-south-1 (Mumbai)</p>
          <!-- Cloud region visualization will be implemented here -->
        </div>
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="font-semibold text-lg mb-2">Azure</h3>
          <p class="text-gray-500">centralindia</p>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="font-semibold text-lg mb-2">GCP</h3>
          <p class="text-gray-500">asia-south1</p>
        </div>
      </div>
    </div>
  `,
})
export class TopologyComponent {}
