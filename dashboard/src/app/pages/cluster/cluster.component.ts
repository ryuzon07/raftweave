import { Component, inject, computed, AfterViewInit, OnDestroy, ElementRef, viewChild } from '@angular/core';
import { MockDataService } from '../../core/services/mock-data.service';
import { StatusPillComponent } from '../../shared/status-pill/status-pill.component';
import { CloudBadgeComponent } from '../../shared/cloud-badge/cloud-badge.component';
import { gsap } from 'gsap';
import { AnimationService } from '../../core/services/animation.service';

@Component({
  selector: 'rw-cluster',
  standalone: true,
  imports: [StatusPillComponent, CloudBadgeComponent],
  template: `
    <div class="cluster" #clusterEl>
      <h1 class="page-title">Cluster Topology</h1>

      <!-- Interactive Cloud Map -->
      <div class="topology-map">
        <div class="cloud-zones">
          @for (cloud of cloudGroups(); track cloud.name) {
            <div class="zone" [class.zone--leader]="cloud.hasLeader" [class]="'zone--' + cloud.name.toLowerCase()">
              <div class="zone__header">
                <span class="zone__icon material-symbols-outlined" [style.color]="cloud.color">cloud</span>
                <span class="zone__name">{{ cloud.name }}</span>
              </div>
              @for (node of cloud.nodes; track node.id) {
                <div class="zone__node" [class.leader]="node.id === topology().leaderId">
                  <!-- Health Ring -->
                  <div class="health-ring" [class]="'ring--' + node.health">
                    <div class="health-ring__inner">
                      @if (node.id === topology().leaderId) {
                        <span class="crown material-symbols-outlined">workspace_premium</span>
                      } @else {
                        <span class="node-dot-lg"></span>
                      }
                    </div>
                  </div>
                  <div class="zone__node-info">
                    <span class="zone__node-region">{{ node.region }}</span>
                    <span class="zone__node-role">{{ node.role }}</span>
                  </div>
                  <div class="zone__node-meta">
                    <rw-status-pill [status]="node.health" />
                    <span class="lag-badge" [class.lag-high]="node.lagMs > 2000">
                      {{ node.lagMs.toFixed(0) }}ms
                    </span>
                  </div>
                </div>
              }
            </div>
          }
        </div>

        <!-- Connection Lines SVG Overlay -->
        <svg class="connections-svg" viewBox="0 0 1000 200" preserveAspectRatio="none" #connectionsSvg>
          <defs>
            <linearGradient id="lineGrad" x1="0" y1="0" x2="1" y2="0">
              <stop offset="0%" stop-color="var(--rw-accent)" stop-opacity="0.3"/>
              <stop offset="50%" stop-color="var(--rw-accent)" stop-opacity="0.8"/>
              <stop offset="100%" stop-color="var(--rw-accent)" stop-opacity="0.3"/>
            </linearGradient>
          </defs>
          <!-- Cross-cloud connection lines -->
          <line x1="280" y1="100" x2="500" y2="100" stroke="url(#lineGrad)" stroke-width="1.5" stroke-dasharray="6 4" class="conn-line" />
          <line x1="500" y1="100" x2="720" y2="100" stroke="url(#lineGrad)" stroke-width="1.5" stroke-dasharray="6 4" class="conn-line" />
          <line x1="280" y1="60" x2="720" y2="140" stroke="url(#lineGrad)" stroke-width="1" stroke-dasharray="4 6" class="conn-line" opacity="0.4" />
          <!-- Data flow dots -->
          <circle r="3" fill="var(--rw-accent)" opacity="0.8">
            <animateMotion dur="3s" repeatCount="indefinite" path="M280,100 L500,100" />
          </circle>
          <circle r="3" fill="var(--rw-accent)" opacity="0.8">
            <animateMotion dur="3.5s" repeatCount="indefinite" path="M500,100 L720,100" />
          </circle>
          <circle r="2" fill="var(--rw-accent)" opacity="0.5">
            <animateMotion dur="4s" repeatCount="indefinite" path="M720,100 L280,100" />
          </circle>
        </svg>
      </div>

      <!-- Per-Region Metrics Table -->
      <div class="section-header">
        <h2 class="section-title">Node Metrics</h2>
      </div>
      <div class="metrics-table">
        <div class="mt-header">
          <span>Region</span><span>Cloud</span><span>Role</span>
          <span>Lag</span><span>Uptime</span><span>Last Ping</span><span>Fencing</span>
        </div>
        @for (node of topology().nodes; track node.id) {
          <div class="mt-row">
            <span class="mt-region">{{ node.region }}</span>
            <rw-cloud-badge [cloud]="node.cloud" [region]="''" />
            <span class="mt-role" [class.leader]="node.role === 'leader'">
              {{ node.role }}
              @if (node.role === 'leader') { <span class="material-symbols-outlined" style="font-size: 14px; vertical-align: middle; margin-left: 2px;">workspace_premium</span> }
            </span>
            <span class="mt-lag" [class.lag-high]="node.lagMs > 2000">
              {{ node.lagMs.toFixed(0) }}ms
            </span>
            <span class="mono">{{ node.uptimePct.toFixed(2) }}%</span>
            <span class="mono time">{{ formatPing(node.lastPing) }}</span>
            <rw-status-pill [status]="node.fencingStatus === 'armed' ? 'healthy' : 'degraded'" [label]="node.fencingStatus" />
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .cluster { max-width: 1400px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }

    /* Topology Map */
    .topology-map {
      position: relative;
      margin-bottom: 32px;
      padding: 24px;
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-lg);
    }
    .cloud-zones {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 20px;
      position: relative;
      z-index: 2;
    }
    .zone {
      background: rgba(66, 68, 127, 0.5);
      border: 1px solid rgba(91, 106, 189, 0.2);
      border-radius: var(--radius-md);
      padding: 16px;
      transition: all var(--transition-base);
    }
    .zone--leader {
      border-color: var(--rw-accent);
      box-shadow: var(--glow-leader);
    }
    .zone__header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 14px;
      padding-bottom: 10px;
      border-bottom: 1px solid rgba(91, 106, 189, 0.15);
    }
    .zone__icon { font-size: 1.3rem; }
    .zone__name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      font-weight: 600;
      color: var(--rw-white);
    }
    .zone__node {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 8px;
      border-radius: var(--radius-sm);
      margin-bottom: 4px;
      transition: background var(--transition-fast);
    }
    .zone__node:hover { background: rgba(91, 106, 189, 0.1); }
    .zone__node.leader {
      background: rgba(119, 126, 240, 0.1);
    }

    /* Health Ring */
    .health-ring {
      width: 36px;
      height: 36px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
      border: 2px solid transparent;
    }
    .ring--healthy { border-color: var(--healthy); }
    .ring--degraded { border-color: var(--degraded); }
    .ring--failed { border-color: var(--failed); }
    .health-ring__inner {
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background: var(--surface-1);
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .crown { font-size: 0.9rem; }
    .node-dot-lg {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--rw-mid);
    }

    .zone__node-info {
      flex: 1;
      display: flex;
      flex-direction: column;
    }
    .zone__node-region {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-white);
      font-weight: 500;
    }
    .zone__node-role {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
    }
    .zone__node-meta {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      gap: 4px;
    }
    .lag-badge {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .lag-badge.lag-high { color: var(--failed); }

    /* Connections SVG */
    .connections-svg {
      position: absolute;
      inset: 0;
      width: 100%;
      height: 100%;
      z-index: 1;
      pointer-events: none;
    }
    .conn-line {
      animation: rw-flow 2s linear infinite;
    }

    /* Metrics Table */
    .section-header { margin-bottom: 16px; }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
    }
    .metrics-table {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-md);
      overflow: hidden;
    }
    .mt-header {
      display: grid;
      grid-template-columns: 1.5fr 1fr 1fr 1fr 1fr 1.5fr 1fr;
      gap: 8px;
      padding: 10px 16px;
      background: rgba(66, 68, 127, 0.5);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .mt-row {
      display: grid;
      grid-template-columns: 1.5fr 1fr 1fr 1fr 1fr 1.5fr 1fr;
      gap: 8px;
      padding: 10px 16px;
      align-items: center;
      border-bottom: 1px solid rgba(91, 106, 189, 0.08);
      transition: background var(--transition-fast);
    }
    .mt-row:hover { background: rgba(91, 106, 189, 0.08); }
    .mt-region {
      font-family: var(--font-mono);
      font-size: var(--text-body);
      color: var(--rw-white);
      font-weight: 500;
    }
    .mt-role {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      text-transform: uppercase;
    }
    .mt-role.leader { color: var(--rw-accent); font-weight: 600; }
    .mt-lag {
      font-family: var(--font-mono);
      font-size: var(--text-body);
      color: var(--rw-mist);
    }
    .mt-lag.lag-high { color: var(--failed); }
    .mono { font-family: var(--font-mono); font-size: var(--text-xs); color: var(--rw-mist); }
    .time { color: var(--rw-mid); }

    @media (max-width: 1024px) {
      .cloud-zones { grid-template-columns: 1fr; }
      .mt-header, .mt-row { grid-template-columns: 1fr 1fr 1fr; }
    }
  `]
})
export class ClusterComponent implements AfterViewInit, OnDestroy {
  mockData = inject(MockDataService);
  private animService = inject(AnimationService);
  private ctx: gsap.Context | undefined;
  clusterEl = viewChild<ElementRef>('clusterEl');
  connectionsSvg = viewChild<ElementRef>('connectionsSvg');

  topology = this.mockData.clusterTopology;

  cloudGroups = computed(() => {
    const nodes = this.topology().nodes;
    const leaderId = this.topology().leaderId;
    return [
      { name: 'AWS', color: '#F59E0B', nodes: nodes.filter(n => n.cloud === 'AWS'), hasLeader: nodes.some(n => n.cloud === 'AWS' && n.id === leaderId) },
      { name: 'Azure', color: '#3B82F6', nodes: nodes.filter(n => n.cloud === 'Azure'), hasLeader: nodes.some(n => n.cloud === 'Azure' && n.id === leaderId) },
      { name: 'GCP', color: '#EF4444', nodes: nodes.filter(n => n.cloud === 'GCP'), hasLeader: nodes.some(n => n.cloud === 'GCP' && n.id === leaderId) },
    ];
  });

  ngAfterViewInit() {
    const el = this.clusterEl();
    if (el) {
      this.ctx = gsap.context(() => {
        this.animService.fadeUpStagger('.zone', el.nativeElement);
        // Line draw-on animation
        const svg = this.connectionsSvg();
        if (svg) {
          const lines = svg.nativeElement.querySelectorAll('.conn-line');
          lines.forEach((line: SVGLineElement) => {
            const length = 300;
            gsap.fromTo(line,
              { strokeDasharray: length, strokeDashoffset: length },
              { strokeDashoffset: 0, duration: 1.5, ease: 'power2.inOut' }
            );
          });
        }
      }, el.nativeElement);
    }
  }

  ngOnDestroy() { this.ctx?.revert(); }

  formatPing(ts: string): string {
    try {
      const diff = Date.now() - new Date(ts).getTime();
      return `${Math.floor(diff / 1000)}s ago`;
    } catch { return '—'; }
  }
}
