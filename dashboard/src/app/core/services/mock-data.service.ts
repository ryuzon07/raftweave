import { Injectable, signal, computed, WritableSignal, Signal } from '@angular/core';
import {
  ClusterTopology, ClusterNode, ClusterEdge, RaftState, RaftNodeState,
  RaftLogEntry, Workload, Build, BuildLogLine, ReplicationLagPoint,
  FailoverEvent, WorkspaceOverview, ReplicationEvent
} from '../models/types';

/**
 * Mock SSE Service — simulates all streaming endpoints with realistic data.
 * All state is exposed via Angular Signals for zoneless reactivity.
 */
@Injectable({ providedIn: 'root' })
export class MockDataService {

  /* ── Workspace Overview ─────────────── */
  readonly workspaceOverview: WritableSignal<WorkspaceOverview> = signal({
    workloadCount: 12,
    clusterHealthPct: 99.7,
    avgLagMs: 342,
    lastFailover: null
  });

  /* ── Cluster Topology ──────────────── */
  private readonly _clusterNodes: WritableSignal<ClusterNode[]> = signal(this._generateClusterNodes());
  private readonly _clusterEdges: WritableSignal<ClusterEdge[]> = signal(this._generateClusterEdges());
  private readonly _leaderId: WritableSignal<string> = signal('node-aws-use1');

  readonly clusterTopology: Signal<ClusterTopology> = computed(() => ({
    nodes: this._clusterNodes(),
    edges: this._clusterEdges(),
    leaderId: this._leaderId(),
    updatedAt: new Date().toISOString()
  }));

  /* ── Raft State ────────────────────── */
  readonly raftState: WritableSignal<RaftState> = signal({
    term: 7,
    leader: 'node-1',
    logIndex: 14832,
    quorumSize: 3,
    electionTimeoutMin: 150,
    electionTimeoutMax: 300,
    nodes: this._generateRaftNodes()
  });

  readonly raftLog: WritableSignal<RaftLogEntry[]> = signal(this._generateRaftLog());

  /* ── Workloads ─────────────────────── */
  readonly workloads: WritableSignal<Workload[]> = signal(this._generateWorkloads());

  /* ── Builds ────────────────────────── */
  readonly builds: WritableSignal<Build[]> = signal(this._generateBuilds());
  readonly activeBuildLogs: WritableSignal<BuildLogLine[]> = signal([]);

  /* ── Replication ───────────────────── */
  readonly replicationLag: WritableSignal<ReplicationLagPoint[]> = signal([]);
  readonly replicationHistory: WritableSignal<ReplicationLagPoint[][]> = signal([]);
  readonly replicationEvents: WritableSignal<ReplicationEvent[]> = signal(this._generateReplicationEvents());

  /* ── Failovers ─────────────────────── */
  readonly failoverEvents: WritableSignal<FailoverEvent[]> = signal(this._generateFailoverEvents());

  /* ── Connection Status ─────────────── */
  readonly connectionStatus: WritableSignal<'connected' | 'reconnecting' | 'disconnected'> = signal('connected');
  readonly healthyNodeCount: WritableSignal<number> = signal(6);

  private _intervals: ReturnType<typeof setInterval>[] = [];

  startStreaming(): void {
    this._startClusterStream();
    this._startRaftStream();
    this._startReplicationStream();
    this._startBuildLogStream();
    this._startFailoverStream();
    this._startOverviewUpdates();
  }

  stopStreaming(): void {
    this._intervals.forEach(id => clearInterval(id));
    this._intervals = [];
  }

  /* ━━━ STREAMING SIMULATORS ━━━ */

  private _startClusterStream(): void {
    const id = setInterval(() => {
      const nodes = this._clusterNodes().map(n => ({
        ...n,
        lagMs: Math.max(0, n.lagMs + (Math.random() - 0.5) * 100),
        lastPing: new Date().toISOString(),
        health: n.lagMs > 5000 ? 'failed' as const :
                n.lagMs > 2000 ? 'degraded' as const : 'healthy' as const
      }));
      this._clusterNodes.set(nodes);
      this.healthyNodeCount.set(nodes.filter(n => n.health === 'healthy').length);
    }, 2000);
    this._intervals.push(id);
  }

  private _startRaftStream(): void {
    const id = setInterval(() => {
      const current = this.raftState();
      const newLogIndex = current.logIndex + Math.floor(Math.random() * 3) + 1;

      // Occasional election simulation (~5% chance)
      if (Math.random() < 0.05) {
        const candidates = current.nodes.filter(n => n.id !== current.leader);
        const newLeader = candidates[Math.floor(Math.random() * candidates.length)];
        this.raftState.set({
          ...current,
          term: current.term + 1,
          leader: newLeader.id,
          logIndex: newLogIndex,
          nodes: current.nodes.map(n => ({
            ...n,
            state: n.id === newLeader.id ? 'leader' as const : 'follower' as const,
            lastHeartbeat: Date.now()
          }))
        });
        this._leaderId.set(
          this._clusterNodes()[Math.floor(Math.random() * this._clusterNodes().length)].id
        );
      } else {
        this.raftState.set({
          ...current,
          logIndex: newLogIndex,
          nodes: current.nodes.map(n => ({
            ...n,
            lastHeartbeat: n.state === 'leader' ? Date.now() :
              Date.now() - Math.floor(Math.random() * 200)
          }))
        });
      }

      // Append log entry
      const entries = this.raftLog();
      const types = ['SET', 'DELETE', 'CONFIG', 'BARRIER', 'NOOP'];
      const newEntry: RaftLogEntry = {
        term: this.raftState().term,
        index: newLogIndex,
        entryType: types[Math.floor(Math.random() * types.length)],
        payload: `key-${Math.random().toString(36).slice(2, 8)}`,
        timestamp: new Date().toISOString()
      };
      this.raftLog.set([newEntry, ...entries].slice(0, 50));
    }, 3000);
    this._intervals.push(id);
  }

  private _startReplicationStream(): void {
    const regions = [
      { region: 'us-east-1', cloud: 'AWS' },
      { region: 'eastus', cloud: 'Azure' },
      { region: 'us-central1', cloud: 'GCP' },
      { region: 'eu-west-1', cloud: 'AWS' },
      { region: 'westeurope', cloud: 'Azure' }
    ];
    const id = setInterval(() => {
      const points = regions.map(r => ({
        ...r,
        lagMs: Math.max(50, 300 + (Math.random() - 0.4) * 600 +
          (Math.random() < 0.03 ? Math.random() * 5000 : 0)),
        timestamp: new Date().toISOString()
      }));
      this.replicationLag.set(points);

      const history = this.replicationHistory();
      const newHistory = [...history, points].slice(-60);
      this.replicationHistory.set(newHistory);
    }, 2000);
    this._intervals.push(id);
  }

  private _startBuildLogStream(): void {
    const logLines = [
      { line: '> Detecting language...', level: 'info' as const },
      { line: '  Detected: Go 1.22', level: 'stdout' as const },
      { line: '> Generating Dockerfile...', level: 'info' as const },
      { line: '  FROM golang:1.22-alpine AS builder', level: 'stdout' as const },
      { line: '  COPY go.mod go.sum ./', level: 'stdout' as const },
      { line: '  RUN go mod download', level: 'stdout' as const },
      { line: '  COPY . .', level: 'stdout' as const },
      { line: '  RUN CGO_ENABLED=0 go build -o /app ./cmd/server', level: 'stdout' as const },
      { line: '> Building with Kaniko...', level: 'info' as const },
      { line: '  [1/8] Resolving base image...', level: 'stdout' as const },
      { line: '  [2/8] Setting up build context...', level: 'stdout' as const },
      { line: '  [3/8] Downloading modules...', level: 'stdout' as const },
      { line: '  [4/8] Compiling packages...', level: 'stdout' as const },
      { line: '  warning: unused import in handler.go:12', level: 'stderr' as const },
      { line: '  [5/8] Linking...', level: 'stdout' as const },
      { line: '  [6/8] Running tests...', level: 'stdout' as const },
      { line: '  PASS: 42 tests passed', level: 'stdout' as const },
      { line: '  [7/8] Creating layer...', level: 'stdout' as const },
      { line: '  [8/8] Packaging image...', level: 'stdout' as const },
      { line: '> Pushing to registry...', level: 'info' as const },
      { line: '  Digest: sha256:a1b2c3d4e5f6...', level: 'stdout' as const },
      { line: '> Build complete ✓', level: 'info' as const },
    ];
    let idx = 0;
    const id = setInterval(() => {
      if (idx < logLines.length) {
        const current = this.activeBuildLogs();
        this.activeBuildLogs.set([...current, {
          ...logLines[idx],
          timestamp: new Date().toISOString()
        }]);
        idx++;
      } else {
        idx = 0;
        this.activeBuildLogs.set([]);
      }
    }, 800);
    this._intervals.push(id);
  }

  private _startFailoverStream(): void {
    const id = setInterval(() => {
      if (Math.random() < 0.1) {
        const triggers: ('node_failure' | 'network_partition' | 'manual')[] =
          ['node_failure', 'network_partition', 'manual'];
        const regions = ['us-east-1', 'eastus', 'us-central1', 'eu-west-1', 'westeurope'];
        const from = regions[Math.floor(Math.random() * regions.length)];
        let to = regions[Math.floor(Math.random() * regions.length)];
        while (to === from) to = regions[Math.floor(Math.random() * regions.length)];

        const event: FailoverEvent = {
          eventId: `fo-${Date.now()}`,
          trigger: triggers[Math.floor(Math.random() * triggers.length)],
          fromRegion: from,
          toRegion: to,
          rtoSeconds: Math.floor(Math.random() * 12) + 2,
          rpoMs: Math.floor(Math.random() * 500),
          term: this.raftState().term,
          status: 'complete',
          occurredAt: new Date().toISOString()
        };
        const events = this.failoverEvents();
        this.failoverEvents.set([event, ...events].slice(0, 50));

        this.workspaceOverview.update(o => ({
          ...o,
          lastFailover: event
        }));
      }
    }, 10000);
    this._intervals.push(id);
  }

  private _startOverviewUpdates(): void {
    const id = setInterval(() => {
      this.workspaceOverview.update(o => ({
        ...o,
        clusterHealthPct: Math.min(100, Math.max(95, o.clusterHealthPct + (Math.random() - 0.5) * 0.4)),
        avgLagMs: Math.max(50, o.avgLagMs + (Math.random() - 0.5) * 80)
      }));
    }, 5000);
    this._intervals.push(id);
  }

  /* ━━━ DATA GENERATORS ━━━ */

  private _generateClusterNodes(): ClusterNode[] {
    return [
      { id: 'node-aws-use1', region: 'us-east-1', cloud: 'AWS', role: 'leader', health: 'healthy', lagMs: 0, uptimePct: 99.99, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
      { id: 'node-aws-euw1', region: 'eu-west-1', cloud: 'AWS', role: 'follower', health: 'healthy', lagMs: 245, uptimePct: 99.95, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
      { id: 'node-azure-eus', region: 'eastus', cloud: 'Azure', role: 'follower', health: 'healthy', lagMs: 312, uptimePct: 99.92, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
      { id: 'node-azure-weu', region: 'westeurope', cloud: 'Azure', role: 'follower', health: 'healthy', lagMs: 189, uptimePct: 99.97, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
      { id: 'node-gcp-usc1', region: 'us-central1', cloud: 'GCP', role: 'follower', health: 'healthy', lagMs: 278, uptimePct: 99.93, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
      { id: 'node-gcp-euw4', region: 'europe-west4', cloud: 'GCP', role: 'follower', health: 'degraded', lagMs: 1820, uptimePct: 98.12, lastPing: new Date().toISOString(), fencingStatus: 'armed' },
    ];
  }

  private _generateClusterEdges(): ClusterEdge[] {
    const nodes = this._generateClusterNodes();
    const edges: ClusterEdge[] = [];
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        edges.push({
          from: nodes[i].id,
          to: nodes[j].id,
          latencyMs: Math.floor(Math.random() * 100) + 20
        });
      }
    }
    return edges;
  }

  private _generateRaftNodes(): RaftNodeState[] {
    return [
      { id: 'node-1', state: 'leader', lastHeartbeat: Date.now() },
      { id: 'node-2', state: 'follower', lastHeartbeat: Date.now() - 50 },
      { id: 'node-3', state: 'follower', lastHeartbeat: Date.now() - 80 },
      { id: 'node-4', state: 'follower', lastHeartbeat: Date.now() - 120 },
      { id: 'node-5', state: 'follower', lastHeartbeat: Date.now() - 30 },
    ];
  }

  private _generateRaftLog(): RaftLogEntry[] {
    const types = ['SET', 'DELETE', 'CONFIG', 'BARRIER', 'NOOP'];
    const entries: RaftLogEntry[] = [];
    for (let i = 0; i < 30; i++) {
      entries.push({
        term: 7,
        index: 14832 - i,
        entryType: types[Math.floor(Math.random() * types.length)],
        payload: `key-${Math.random().toString(36).slice(2, 8)}`,
        timestamp: new Date(Date.now() - i * 3000).toISOString()
      });
    }
    return entries;
  }

  private _generateWorkloads(): Workload[] {
    return [
      {
        id: 'wl-1', name: 'api-gateway', language: 'Go',
        regions: [
          { region: 'us-east-1', cloud: 'AWS', status: 'active' },
          { region: 'eastus', cloud: 'Azure', status: 'active' },
          { region: 'us-central1', cloud: 'GCP', status: 'standby' },
        ],
        health: 'healthy', lastDeploy: '2 hours ago',
        descriptorYaml: 'name: api-gateway\nruntime: go\nreplicas: 3\nport: 8080\nhealth_check: /healthz',
        envVars: [
          { key: 'DATABASE_URL', value: 'postgresql://...', masked: true },
          { key: 'LOG_LEVEL', value: 'info', masked: false },
          { key: 'API_KEY', value: 'rw_live_...', masked: true },
        ]
      },
      {
        id: 'wl-2', name: 'consensus-engine', language: 'Rust',
        regions: [
          { region: 'us-east-1', cloud: 'AWS', status: 'active' },
          { region: 'eu-west-1', cloud: 'AWS', status: 'active' },
        ],
        health: 'healthy', lastDeploy: '45 minutes ago',
        descriptorYaml: 'name: consensus-engine\nruntime: rust\nreplicas: 5\nport: 9090',
        envVars: [
          { key: 'RAFT_CLUSTER_ID', value: 'cluster-prod-01', masked: false },
          { key: 'ELECTION_TIMEOUT', value: '300', masked: false },
        ]
      },
      {
        id: 'wl-3', name: 'state-sync', language: 'TypeScript',
        regions: [
          { region: 'us-central1', cloud: 'GCP', status: 'active' },
          { region: 'westeurope', cloud: 'Azure', status: 'deploying' },
        ],
        health: 'degraded', lastDeploy: '12 minutes ago',
        descriptorYaml: 'name: state-sync\nruntime: node\nreplicas: 2\nport: 3000',
        envVars: [
          { key: 'REDIS_URL', value: 'redis://...', masked: true },
        ]
      },
      {
        id: 'wl-4', name: 'edge-proxy', language: 'Go',
        regions: [
          { region: 'us-east-1', cloud: 'AWS', status: 'active' },
          { region: 'eastus', cloud: 'Azure', status: 'active' },
          { region: 'us-central1', cloud: 'GCP', status: 'active' },
          { region: 'eu-west-1', cloud: 'AWS', status: 'active' },
        ],
        health: 'healthy', lastDeploy: '6 hours ago',
        descriptorYaml: 'name: edge-proxy\nruntime: go\nreplicas: 4\nport: 443',
        envVars: []
      },
      {
        id: 'wl-5', name: 'metrics-collector', language: 'Python',
        regions: [
          { region: 'us-east-1', cloud: 'AWS', status: 'active' },
        ],
        health: 'healthy', lastDeploy: '1 day ago',
        descriptorYaml: 'name: metrics-collector\nruntime: python\nreplicas: 1\nport: 9100',
        envVars: [
          { key: 'PROMETHEUS_URL', value: 'http://...', masked: false },
        ]
      },
      {
        id: 'wl-6', name: 'notification-service', language: 'Java',
        regions: [
          { region: 'eastus', cloud: 'Azure', status: 'active' },
          { region: 'westeurope', cloud: 'Azure', status: 'active' },
        ],
        health: 'failed', lastDeploy: '3 days ago',
        descriptorYaml: 'name: notification-service\nruntime: java\nreplicas: 2\nport: 8081',
        envVars: [
          { key: 'SMTP_HOST', value: 'smtp.raftweave.io', masked: false },
          { key: 'SMTP_PASSWORD', value: '***', masked: true },
        ]
      },
    ];
  }

  private _generateBuilds(): Build[] {
    return [
      {
        id: 'build-001', workloadId: 'wl-1', workloadName: 'api-gateway',
        branch: 'main', commitSha: 'a1b2c3d', trigger: 'webhook', language: 'Go',
        status: 'building',
        steps: [
          { name: 'Detect', status: 'complete', elapsed: '1.2s', icon: 'search' },
          { name: 'Dockerfile', status: 'complete', elapsed: '0.8s', icon: 'description' },
          { name: 'Build', status: 'running', elapsed: '34s', icon: 'build' },
          { name: 'Push', status: 'pending', elapsed: '—', icon: 'upload' },
          { name: 'Done', status: 'pending', elapsed: '—', icon: 'check_circle' },
        ],
        duration: '36s', imageDigest: '', timestamp: new Date().toISOString()
      },
      {
        id: 'build-002', workloadId: 'wl-2', workloadName: 'consensus-engine',
        branch: 'main', commitSha: 'f4e5d6c', trigger: 'manual', language: 'Rust',
        status: 'complete',
        steps: [
          { name: 'Detect', status: 'complete', elapsed: '0.9s', icon: 'search' },
          { name: 'Dockerfile', status: 'complete', elapsed: '1.1s', icon: 'description' },
          { name: 'Build', status: 'complete', elapsed: '2m 12s', icon: 'build' },
          { name: 'Push', status: 'complete', elapsed: '8s', icon: 'upload' },
          { name: 'Done', status: 'complete', elapsed: '', icon: 'check_circle' },
        ],
        duration: '2m 22s', imageDigest: 'sha256:f4e5d6c7a8b9', timestamp: new Date(Date.now() - 3600000).toISOString()
      },
      {
        id: 'build-003', workloadId: 'wl-6', workloadName: 'notification-service',
        branch: 'fix/smtp-timeout', commitSha: '9c8b7a6', trigger: 'webhook', language: 'Java',
        status: 'failed',
        steps: [
          { name: 'Detect', status: 'complete', elapsed: '1.0s', icon: 'search' },
          { name: 'Dockerfile', status: 'complete', elapsed: '0.7s', icon: 'description' },
          { name: 'Build', status: 'failed', elapsed: '45s', icon: 'build' },
          { name: 'Push', status: 'pending', elapsed: '—', icon: 'upload' },
          { name: 'Done', status: 'pending', elapsed: '—', icon: 'check_circle' },
        ],
        duration: '47s', imageDigest: '', timestamp: new Date(Date.now() - 7200000).toISOString()
      },
    ];
  }

  private _generateReplicationEvents(): ReplicationEvent[] {
    return [
      { timestamp: new Date(Date.now() - 300000).toISOString(), region: 'eu-west-1', eventType: 'LAG_SPIKE', lagAtEvent: 4200, actionTaken: 'Alert triggered' },
      { timestamp: new Date(Date.now() - 600000).toISOString(), region: 'westeurope', eventType: 'CATCHUP_COMPLETE', lagAtEvent: 120, actionTaken: 'None' },
      { timestamp: new Date(Date.now() - 1200000).toISOString(), region: 'us-central1', eventType: 'SYNC_PAUSE', lagAtEvent: 890, actionTaken: 'Throttled writes' },
      { timestamp: new Date(Date.now() - 3600000).toISOString(), region: 'eastus', eventType: 'LAG_SPIKE', lagAtEvent: 6100, actionTaken: 'Fencing triggered' },
    ];
  }

  private _generateFailoverEvents(): FailoverEvent[] {
    return [
      { eventId: 'fo-001', trigger: 'node_failure', fromRegion: 'us-east-1', toRegion: 'eastus', rtoSeconds: 4, rpoMs: 120, term: 6, status: 'complete', occurredAt: new Date(Date.now() - 86400000).toISOString() },
      { eventId: 'fo-002', trigger: 'network_partition', fromRegion: 'eastus', toRegion: 'us-central1', rtoSeconds: 7, rpoMs: 340, term: 5, status: 'complete', occurredAt: new Date(Date.now() - 172800000).toISOString() },
      { eventId: 'fo-003', trigger: 'manual', fromRegion: 'us-central1', toRegion: 'us-east-1', rtoSeconds: 3, rpoMs: 0, term: 4, status: 'complete', occurredAt: new Date(Date.now() - 432000000).toISOString() },
      { eventId: 'fo-004', trigger: 'node_failure', fromRegion: 'eu-west-1', toRegion: 'westeurope', rtoSeconds: 9, rpoMs: 510, term: 3, status: 'complete', occurredAt: new Date(Date.now() - 604800000).toISOString() },
      { eventId: 'fo-005', trigger: 'network_partition', fromRegion: 'westeurope', toRegion: 'eu-west-1', rtoSeconds: 5, rpoMs: 200, term: 2, status: 'rolled_back', occurredAt: new Date(Date.now() - 864000000).toISOString() },
    ];
  }
}
