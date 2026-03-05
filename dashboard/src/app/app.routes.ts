import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'topology', pathMatch: 'full' },
  {
    path: 'onboarding',
    loadComponent: () =>
      import('./features/onboarding/onboarding.component').then(
        (m) => m.OnboardingComponent
      ),
  },
  {
    path: 'topology',
    loadComponent: () =>
      import('./features/topology/topology.component').then(
        (m) => m.TopologyComponent
      ),
  },
  {
    path: 'raft',
    loadComponent: () =>
      import('./features/raft-visualizer/raft-visualizer.component').then(
        (m) => m.RaftVisualizerComponent
      ),
  },
  {
    path: 'builds',
    loadComponent: () =>
      import('./features/build-pipeline/build-pipeline.component').then(
        (m) => m.BuildPipelineComponent
      ),
  },
  {
    path: 'replication',
    loadComponent: () =>
      import('./features/replication/replication.component').then(
        (m) => m.ReplicationComponent
      ),
  },
  {
    path: 'failover-log',
    loadComponent: () =>
      import('./features/failover-log/failover-log.component').then(
        (m) => m.FailoverLogComponent
      ),
  },
];
