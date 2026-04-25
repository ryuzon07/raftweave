/* ━━━ DATA MODELS ━━━ */

export interface ClusterNode {
  id: string;
  region: string;
  cloud: 'AWS' | 'Azure' | 'GCP';
  role: 'leader' | 'follower' | 'candidate';
  health: 'healthy' | 'degraded' | 'failed';
  lagMs: number;
  uptimePct: number;
  lastPing: string;
  fencingStatus: 'armed' | 'disarmed' | 'triggered';
}

export interface ClusterEdge {
  from: string;
  to: string;
  latencyMs: number;
}

export interface ClusterTopology {
  nodes: ClusterNode[];
  edges: ClusterEdge[];
  leaderId: string;
  updatedAt: string;
}

export interface RaftNodeState {
  id: string;
  state: 'leader' | 'follower' | 'candidate';
  lastHeartbeat: number;
}

export interface RaftState {
  term: number;
  leader: string;
  logIndex: number;
  quorumSize: number;
  electionTimeoutMin: number;
  electionTimeoutMax: number;
  nodes: RaftNodeState[];
}

export interface RaftLogEntry {
  term: number;
  index: number;
  entryType: string;
  payload: string;
  timestamp: string;
}

export interface Workload {
  id: string;
  name: string;
  language: string;
  regions: WorkloadRegion[];
  health: 'healthy' | 'degraded' | 'failed';
  lastDeploy: string;
  descriptorYaml: string;
  envVars: { key: string; value: string; masked: boolean }[];
}

export interface WorkloadRegion {
  region: string;
  cloud: 'AWS' | 'Azure' | 'GCP';
  status: 'active' | 'standby' | 'deploying';
}

export interface BuildStep {
  name: string;
  status: 'pending' | 'running' | 'complete' | 'failed';
  elapsed: string;
  icon: string;
}

export interface Build {
  id: string;
  workloadId: string;
  workloadName: string;
  branch: string;
  commitSha: string;
  trigger: 'webhook' | 'manual';
  language: string;
  status: 'building' | 'complete' | 'failed';
  steps: BuildStep[];
  duration: string;
  imageDigest: string;
  timestamp: string;
}

export interface BuildLogLine {
  line: string;
  level: 'stdout' | 'stderr' | 'info';
  timestamp: string;
}

export interface ReplicationLagPoint {
  region: string;
  cloud: string;
  lagMs: number;
  timestamp: string;
}

export interface ReplicationEvent {
  timestamp: string;
  region: string;
  eventType: string;
  lagAtEvent: number;
  actionTaken: string;
}

export interface FailoverEvent {
  eventId: string;
  trigger: 'node_failure' | 'network_partition' | 'manual';
  fromRegion: string;
  toRegion: string;
  rtoSeconds: number;
  rpoMs: number;
  term: number;
  status: 'complete' | 'in_progress' | 'rolled_back';
  occurredAt: string;
}

export interface WorkspaceOverview {
  workloadCount: number;
  clusterHealthPct: number;
  avgLagMs: number;
  lastFailover: FailoverEvent | null;
}

export type HealthStatus = 'healthy' | 'degraded' | 'failed' | 'pending';
export type CloudProvider = 'AWS' | 'Azure' | 'GCP';
