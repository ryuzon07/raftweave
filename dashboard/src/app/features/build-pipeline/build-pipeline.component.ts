import { Component } from '@angular/core';

@Component({
  selector: 'app-build-pipeline',
  standalone: true,
  template: `
    <div>
      <h2 class="text-2xl font-bold mb-6">Build Pipeline</h2>
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-gray-600">Build job monitoring and log streaming will be displayed here.</p>
        <!-- Build job list and streaming logs will be implemented here -->
      </div>
    </div>
  `,
})
export class BuildPipelineComponent {}
