import { Component, inject } from '@angular/core';
import { clusterState } from '../../shared/signals/cluster-state.signal';

@Component({
  selector: 'app-raft-visualizer',
  standalone: true,
  template: `
    <div>
      <h2 class="text-2xl font-bold mb-6">Raft Consensus Cluster</h2>
      <div class="bg-white rounded-lg shadow p-6 mb-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-semibold text-lg">Cluster State</h3>
          <span class="text-sm text-gray-500">Term: {{ state().currentTerm }}</span>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <!-- Raft node visualizations will be rendered here -->
        </div>
      </div>
    </div>
  `,
})
export class RaftVisualizerComponent {
  state = clusterState;
}
