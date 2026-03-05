import { Component } from '@angular/core';

@Component({
  selector: 'app-failover-log',
  standalone: true,
  template: `
    <div>
      <h2 class="text-2xl font-bold mb-6">Failover Log</h2>
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Workload</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">From</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">To</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">RTO</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">RPO</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <!-- Failover event rows will be rendered here -->
          </tbody>
        </table>
      </div>
    </div>
  `,
})
export class FailoverLogComponent {}
