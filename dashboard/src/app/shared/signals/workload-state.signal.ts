import { signal } from '@angular/core';

export interface WorkloadInfo {
  id: string;
  name: string;
  status: 'PENDING' | 'DEPLOYING' | 'RUNNING' | 'FAILING_OVER' | 'DEGRADED' | 'STOPPED';
  primaryCloud: string;
  primaryRegion: string;
  replicaCloud?: string;
  replicaRegion?: string;
  lastDeployedAt?: string;
}

export const workloadState = signal<WorkloadInfo[]>([]);
