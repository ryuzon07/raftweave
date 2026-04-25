import { Component } from '@angular/core';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  template: `
    <aside class="w-64 bg-gray-900 text-white flex flex-col">
      <div class="p-4 text-xl font-bold border-b border-gray-700">
        RaftWeave
      </div>
      <nav class="flex-1 p-4 space-y-2">
        <a routerLink="/topology" class="block px-3 py-2 rounded hover:bg-gray-700">Topology</a>
        <a routerLink="/raft" class="block px-3 py-2 rounded hover:bg-gray-700">Raft Cluster</a>
        <a routerLink="/builds" class="block px-3 py-2 rounded hover:bg-gray-700">Build Pipeline</a>
        <a routerLink="/replication" class="block px-3 py-2 rounded hover:bg-gray-700">Replication</a>
        <a routerLink="/failover-log" class="block px-3 py-2 rounded hover:bg-gray-700">Failover Log</a>
        <a routerLink="/onboarding" class="block px-3 py-2 rounded hover:bg-gray-700">Onboarding</a>
      </nav>
    </aside>
  `,
})
export class SidebarComponent {}
