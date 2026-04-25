import { signal } from '@angular/core';

export interface RaftNodeState {
  nodeId: string;
  role: 'LEADER' | 'FOLLOWER' | 'CANDIDATE';
  term: number;
  lastLogIndex: number;
  commitIndex: number;
  address: string;
}

export interface ClusterState {
  nodes: RaftNodeState[];
  leaderId: string;
  currentTerm: number;
  clusterHealthy: boolean;
}

export const clusterState = signal<ClusterState>({
  nodes: [],
  leaderId: '',
  currentTerm: 0,
  clusterHealthy: false,
});
