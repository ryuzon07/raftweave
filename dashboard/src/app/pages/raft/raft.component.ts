import { Component, inject, computed, AfterViewInit, OnDestroy, ElementRef, viewChild, effect } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { AnimationService } from '../../core/services/animation.service';
import { ToastService } from '../../core/services/toast.service';
import { gsap } from 'gsap';

@Component({
  selector: 'rw-raft',
  standalone: true,
  imports: [],
  template: `
    <div class="raft" #raftEl>
      <h1 class="page-title">Raft Engine</h1>

      <div class="raft-split">
        <!-- Node Graph -->
        <div class="raft-graph">
          <div class="section-header">
            <h2 class="section-title">Node State Visualizer</h2>
          </div>
          <div class="node-graph-container">
            <svg class="node-graph" viewBox="0 0 400 400" #nodeSvg>
              <defs>
                <filter id="leaderGlow">
                  <feGaussianBlur stdDeviation="8" result="blur"/>
                  <feComposite in="SourceGraphic" in2="blur" operator="over"/>
                </filter>
              </defs>
              <!-- Edges (heartbeat lines) -->
              @for (edge of nodeEdges(); track edge.from + edge.to) {
                <line
                  [attr.x1]="edge.x1" [attr.y1]="edge.y1"
                  [attr.x2]="edge.x2" [attr.y2]="edge.y2"
                  stroke="var(--rw-mid)" stroke-width="1" stroke-dasharray="4 4"
                  opacity="0.4"
                />
                <!-- Heartbeat direction arrow -->
                @if (edge.isHeartbeat) {
                  <circle r="3" fill="var(--rw-accent)" opacity="0.7">
                    <animateMotion [attr.dur]="1.5 + $index * 0.3 + 's'" repeatCount="indefinite"
                      [attr.path]="'M' + edge.x1 + ',' + edge.y1 + ' L' + edge.x2 + ',' + edge.y2" />
                  </circle>
                }
              }
              <!-- Nodes -->
              @for (node of nodePositions(); track node.id; let i = $index) {
                <g class="raft-node" [class]="'raft-node--' + node.state" #nodeEl>
                  <!-- Outer glow for leader -->
                  @if (node.state === 'leader') {
                    <circle [attr.cx]="node.x" [attr.cy]="node.y" r="38"
                      fill="none" stroke="var(--rw-accent)" stroke-width="2"
                      opacity="0.3" filter="url(#leaderGlow)">
                      <animate attributeName="r" values="38;42;38" dur="2s" repeatCount="indefinite"/>
                    </circle>
                  }
                  <!-- Node circle -->
                  <circle [attr.cx]="node.x" [attr.cy]="node.y" r="30"
                    [attr.fill]="node.state === 'leader' ? 'var(--rw-accent)' : node.state === 'candidate' ? 'var(--degraded)' : 'var(--surface-3)'"
                    stroke="var(--rw-mid)" stroke-width="1.5" />
                  <!-- Node label -->
                  <text [attr.x]="node.x" [attr.y]="node.y - 4"
                    text-anchor="middle" fill="var(--rw-white)"
                    font-family="var(--font-mono)" font-size="10" font-weight="500">
                    {{ node.id }}
                  </text>
                  <!-- State label -->
                  <text [attr.x]="node.x" [attr.y]="node.y + 10"
                    text-anchor="middle"
                    [attr.fill]="node.state === 'leader' ? 'var(--rw-white)' : 'var(--rw-mist)'"
                    font-family="var(--font-mono)" font-size="7" text-transform="uppercase">
                    {{ node.state }}
                  </text>
                  <!-- Crown -->
                  @if (node.state === 'leader') {
                    <text [attr.x]="node.x" [attr.y]="node.y - 36"
                      text-anchor="middle" font-size="16" class="material-symbols-outlined" fill="var(--rw-white)">workspace_premium</text>
                  }
                </g>
              }
            </svg>
          </div>
        </div>

        <!-- Live Raft Metrics -->
        <div class="raft-metrics">
          <div class="section-header">
            <h2 class="section-title">Live Metrics</h2>
          </div>

          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-label">Current Term</span>
              <span class="stat-value accent">{{ raftState().term }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Leader Node</span>
              <span class="stat-value">{{ raftState().leader }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Log Index</span>
              <span class="stat-value mono">{{ raftState().logIndex }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Quorum Size</span>
              <span class="stat-value">{{ raftState().quorumSize }}/{{ raftState().nodes.length }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Election Timeout</span>
              <span class="stat-value mono">{{ raftState().electionTimeoutMin }}–{{ raftState().electionTimeoutMax }}ms</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Last Heartbeat</span>
              <span class="stat-value mono">{{ lastHeartbeatMs() }}ms ago</span>
            </div>
          </div>

          <!-- Raft Log Tail -->
          <div class="section-header" style="margin-top: 24px;">
            <h2 class="section-title">Raft Log</h2>
          </div>
          <div class="raft-log rw-terminal">
            @for (entry of raftLog(); track entry.index) {
              <div class="log-entry">
                <span class="log-meta">[Term:{{ entry.term }}][Idx:{{ entry.index }}]</span>
                <span class="log-type" [class]="'type--' + entry.entryType.toLowerCase()">{{ entry.entryType }}</span>
                <span class="log-payload">— {{ entry.payload }}</span>
              </div>
            }
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .raft { max-width: 1400px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }
    .raft-split {
      display: grid;
      grid-template-columns: 7fr 5fr;
      gap: 24px;
    }

    /* Node Graph */
    .raft-graph {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-lg);
      padding: 20px;
    }
    .node-graph-container {
      display: flex;
      justify-content: center;
      padding: 20px 0;
    }
    .node-graph {
      width: 100%;
      max-width: 400px;
      height: auto;
    }

    /* Metrics */
    .raft-metrics {
      display: flex;
      flex-direction: column;
    }
    .section-header { margin-bottom: 12px; }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
    }
    .stats-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 10px;
    }
    .stat-item {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      padding: 14px;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }
    .stat-label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .stat-value {
      font-family: var(--font-heading);
      font-size: 1.2rem;
      font-weight: 600;
      color: var(--rw-white);
    }
    .stat-value.accent { color: var(--rw-accent); }
    .stat-value.mono { font-family: var(--font-mono); font-size: 1rem; }

    /* Raft Log */
    .raft-log {
      flex: 1;
      min-height: 200px;
      max-height: 350px;
      overflow-y: auto;
    }
    .log-entry {
      display: flex;
      gap: 6px;
      padding: 2px 0;
      font-size: var(--text-xs);
    }
    .log-meta {
      color: var(--rw-mid);
      font-family: var(--font-mono);
      white-space: nowrap;
    }
    .log-type {
      font-family: var(--font-mono);
      font-weight: 500;
      white-space: nowrap;
    }
    .type--set { color: var(--healthy); }
    .type--delete { color: var(--failed); }
    .type--config { color: var(--rw-accent); }
    .type--barrier { color: var(--degraded); }
    .type--noop { color: var(--rw-mid); }
    .log-payload {
      color: var(--rw-mist);
      font-family: var(--font-mono);
    }

    @media (max-width: 1024px) {
      .raft-split { grid-template-columns: 1fr; }
    }
  `]
})
export class RaftComponent implements AfterViewInit, OnDestroy {
  mockData = inject(MockDataService);
  private animService = inject(AnimationService);
  private toastService = inject(ToastService);
  private ctx: gsap.Context | undefined;

  raftEl = viewChild<ElementRef>('raftEl');

  raftState = this.mockData.raftState;
  raftLog = this.mockData.raftLog;

  private prevTerm = this.raftState().term;

  // Track elections and show toast
  private electionWatcher = effect(() => {
    const state = this.raftState();
    if (state.term > this.prevTerm) {
      this.toastService.election(
        `Leader Election — Term ${state.term}`,
        `New Leader: ${state.leader}`
      );
      this.prevTerm = state.term;
    }
  });

  lastHeartbeatMs = computed(() => {
    const nodes = this.raftState().nodes;
    const leader = nodes.find(n => n.state === 'leader');
    if (!leader) return 0;
    return Date.now() - leader.lastHeartbeat;
  });

  // 5 nodes in a circle
  nodePositions = computed(() => {
    const nodes = this.raftState().nodes;
    const cx = 200, cy = 200, r = 120;
    return nodes.map((n, i) => {
      const angle = (i / nodes.length) * Math.PI * 2 - Math.PI / 2;
      return {
        ...n,
        x: cx + Math.cos(angle) * r,
        y: cy + Math.sin(angle) * r
      };
    });
  });

  nodeEdges = computed(() => {
    const positions = this.nodePositions();
    const leader = positions.find(n => n.state === 'leader');
    const edges: any[] = [];
    for (let i = 0; i < positions.length; i++) {
      for (let j = i + 1; j < positions.length; j++) {
        edges.push({
          from: positions[i].id,
          to: positions[j].id,
          x1: positions[i].x,
          y1: positions[i].y,
          x2: positions[j].x,
          y2: positions[j].y,
          isHeartbeat: leader && (positions[i].id === leader.id || positions[j].id === leader.id)
        });
      }
    }
    return edges;
  });

  ngAfterViewInit() {
    const el = this.raftEl();
    if (el) {
      this.ctx = gsap.context(() => {
        this.animService.fadeUpStagger('.stat-item', el.nativeElement);
      }, el.nativeElement);
    }
  }

  ngOnDestroy() { this.ctx?.revert(); }
}
