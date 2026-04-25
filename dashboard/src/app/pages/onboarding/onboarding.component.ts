import { Component, signal } from '@angular/core';

@Component({
  selector: 'rw-onboarding',
  standalone: true,
  template: `
    <div class="onboarding">
      <h1 class="page-title">Deploy New Workload</h1>

      <!-- Progress Bar -->
      <div class="progress-bar">
        @for (step of steps; track step.number) {
          <div class="progress-step" [class.active]="currentStep() >= step.number" [class.current]="currentStep() === step.number">
            <div class="progress-step__num">{{ step.number }}</div>
            <span class="progress-step__label">{{ step.label }}</span>
          </div>
          @if (!$last) {
            <div class="progress-connector" [class.active]="currentStep() > step.number"></div>
          }
        }
      </div>

      <!-- Step Content -->
      <div class="step-content">
        @switch (currentStep()) {
          @case (1) {
            <div class="step">
              <h2 class="step-title">Workload Descriptor</h2>
              <p class="step-desc">Define your workload configuration in YAML format.</p>
              <div class="editor-area">
                <textarea class="yaml-editor rw-terminal" rows="16"
                  [value]="yamlContent()"
                  (input)="yamlContent.set(getInputValue($event))">
                </textarea>
                <div class="editor-help">
                  <h4 class="help-title">Field Reference</h4>
                  <div class="help-item"><span class="help-key">name</span> Workload identifier</div>
                  <div class="help-item"><span class="help-key">runtime</span> go | rust | node | python | java</div>
                  <div class="help-item"><span class="help-key">replicas</span> Number of instances</div>
                  <div class="help-item"><span class="help-key">port</span> Container port</div>
                  <div class="help-item"><span class="help-key">health_check</span> Health endpoint path</div>
                </div>
              </div>
            </div>
          }
          @case (2) {
            <div class="step">
              <h2 class="step-title">Cloud Credentials</h2>
              <p class="step-desc">Provide credentials for each cloud provider you want to deploy to.</p>
              <div class="cred-tabs">
                @for (tab of cloudTabs; track tab.name) {
                  <button class="cred-tab" [class.active]="activeCloud() === tab.name" (click)="activeCloud.set(tab.name)">
                    <span class="material-symbols-outlined" [style.color]="tab.color" style="font-size: 18px; vertical-align: text-bottom;">cloud</span> {{ tab.name }}
                  </button>
                }
              </div>
              <div class="cred-form">
                @switch (activeCloud()) {
                  @case ('AWS') {
                    <div class="form-group">
                      <label class="form-label">IAM Role ARN</label>
                      <input class="form-input" type="text" placeholder="arn:aws:iam::123456789:role/raftweave" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">Region</label>
                      <select class="form-input"><option>us-east-1</option><option>eu-west-1</option><option>ap-southeast-1</option></select>
                    </div>
                  }
                  @case ('Azure') {
                    <div class="form-group">
                      <label class="form-label">Tenant ID</label>
                      <input class="form-input" type="text" placeholder="xxxxxxxx-xxxx-xxxx" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">Client ID</label>
                      <input class="form-input" type="text" placeholder="xxxxxxxx-xxxx-xxxx" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">Client Secret</label>
                      <input class="form-input" type="password" placeholder="••••••••" />
                    </div>
                  }
                  @case ('GCP') {
                    <div class="form-group">
                      <label class="form-label">Service Account JSON</label>
                      <div class="file-upload">
                        <span class="file-upload__icon material-symbols-outlined">folder</span>
                        <span class="file-upload__text">Drop JSON key file or click to upload</span>
                      </div>
                    </div>
                  }
                }
                <div class="encryption-notice"><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: text-bottom; margin-right: 4px;">lock</span> Credentials encrypted at rest with AES-256-GCM</div>
              </div>
            </div>
          }
          @case (3) {
            <div class="step">
              <h2 class="step-title">Target Regions</h2>
              <p class="step-desc">Select deployment regions across cloud providers.</p>
              <div class="region-picker">
                @for (region of availableRegions; track region.id) {
                  <label class="region-item" [class.selected]="selectedRegions().includes(region.id)">
                    <input type="checkbox" [checked]="selectedRegions().includes(region.id)"
                      (change)="toggleRegion(region.id)" class="region-checkbox" />
                    <span class="region-icon material-symbols-outlined" [style.color]="region.color" style="font-size: 18px; vertical-align: text-bottom;">cloud</span>
                    <span class="region-name">{{ region.name }}</span>
                    <span class="region-cloud">{{ region.cloud }}</span>
                  </label>
                }
              </div>
              <div class="selected-summary">
                <h4>Selected: {{ selectedRegions().length }} regions</h4>
              </div>
            </div>
          }
          @case (4) {
            <div class="step">
              <h2 class="step-title">Review & Deploy</h2>
              <p class="step-desc">Review your configuration before deploying.</p>
              <div class="review-card rw-card-static">
                <div class="review-row"><span class="review-label">Workload</span><span class="review-value">{{ parseName() }}</span></div>
                <div class="review-row"><span class="review-label">Regions</span><span class="review-value">{{ selectedRegions().length }} selected</span></div>
                <div class="review-row"><span class="review-label">Clouds</span><span class="review-value">{{ uniqueClouds() }}</span></div>
              </div>
              <button class="deploy-btn rw-btn-primary" [class.deploying]="deploying()" (click)="deploy()">
                @if (deploying()) {
                  <span class="spinner"></span> Deploying...
                } @else if (deployed()) {
                  <span class="material-symbols-outlined" style="vertical-align: text-bottom; margin-right: 4px;">check_circle</span> Deployed Successfully!
                } @else {
                  <span class="material-symbols-outlined" style="vertical-align: text-bottom; margin-right: 4px;">rocket_launch</span> Deploy Workload
                }
              </button>
            </div>
          }
        }
      </div>

      <!-- Navigation -->
      <div class="step-nav">
        @if (currentStep() > 1) {
          <button class="rw-btn-ghost" (click)="currentStep.set(currentStep() - 1)">← Back</button>
        }
        @if (currentStep() < 4) {
          <button class="rw-btn-primary" (click)="currentStep.set(currentStep() + 1)">Next →</button>
        }
      </div>
    </div>
  `,
  styles: [`
    .onboarding { max-width: 900px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }

    /* Progress Bar */
    .progress-bar {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0;
      margin-bottom: 36px;
    }
    .progress-step {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 6px;
    }
    .progress-step__num {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: var(--surface-2);
      border: 2px solid var(--rw-mid);
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--font-heading);
      font-size: var(--text-body);
      font-weight: 600;
      color: var(--rw-mid);
      transition: all var(--transition-base);
    }
    .progress-step.active .progress-step__num {
      border-color: var(--rw-accent);
      background: var(--rw-accent);
      color: var(--rw-white);
    }
    .progress-step.current .progress-step__num {
      box-shadow: var(--glow-accent);
    }
    .progress-step__label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .progress-step.active .progress-step__label { color: var(--rw-mist); }
    .progress-connector {
      width: 60px;
      height: 2px;
      background: var(--rw-mid);
      opacity: 0.3;
      margin: 0 8px;
      margin-bottom: 20px;
    }
    .progress-connector.active { background: var(--rw-accent); opacity: 1; }

    /* Steps */
    .step-content {
      background: var(--surface-2);
      border: 1px solid rgba(91, 106, 189, 0.15);
      border-radius: var(--radius-lg);
      padding: 28px;
      margin-bottom: 20px;
      min-height: 400px;
    }
    .step-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
      margin-bottom: 8px;
    }
    .step-desc {
      font-size: var(--text-body);
      color: var(--rw-mist);
      margin-bottom: 20px;
    }

    /* YAML Editor */
    .editor-area { display: grid; grid-template-columns: 1fr 250px; gap: 16px; }
    .yaml-editor {
      width: 100%;
      resize: vertical;
      font-size: var(--text-mono);
      line-height: 1.6;
      border: 1px solid rgba(91, 106, 189, 0.2);
      outline: none;
    }
    .yaml-editor:focus { border-color: var(--rw-accent); }
    .editor-help {
      background: rgba(66, 68, 127, 0.3);
      border-radius: var(--radius-sm);
      padding: 14px;
    }
    .help-title {
      font-family: var(--font-heading);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 10px;
    }
    .help-item {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      padding: 4px 0;
    }
    .help-key {
      color: var(--rw-accent);
      margin-right: 6px;
    }

    /* Credentials */
    .cred-tabs { display: flex; gap: 8px; margin-bottom: 20px; }
    .cred-tab {
      padding: 8px 16px;
      background: var(--surface-1);
      border: 1px solid rgba(91, 106, 189, 0.2);
      color: var(--rw-mist);
      border-radius: var(--radius-sm);
      font-family: var(--font-body);
      font-size: var(--text-body);
      cursor: pointer;
      transition: all var(--transition-fast);
    }
    .cred-tab.active {
      border-color: var(--rw-accent);
      color: var(--rw-white);
      background: rgba(119, 126, 240, 0.12);
    }
    .cred-form { max-width: 500px; }
    .form-group { margin-bottom: 16px; }
    .form-label {
      display: block;
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--rw-mist);
      margin-bottom: 6px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .form-input {
      width: 100%;
      padding: 10px 14px;
      background: var(--surface-1);
      border: 1px solid rgba(91, 106, 189, 0.3);
      border-radius: var(--radius-sm);
      color: var(--rw-white);
      font-family: var(--font-mono);
      font-size: var(--text-body);
      outline: none;
      transition: border-color var(--transition-fast);
    }
    .form-input:focus { border-color: var(--rw-accent); }
    .file-upload {
      border: 2px dashed rgba(91, 106, 189, 0.3);
      border-radius: var(--radius-md);
      padding: 32px;
      text-align: center;
      cursor: pointer;
      transition: all var(--transition-fast);
    }
    .file-upload:hover { border-color: var(--rw-accent); }
    .file-upload__icon { font-size: 2rem; display: block; margin-bottom: 8px; }
    .file-upload__text { color: var(--rw-mist); font-size: var(--text-body); }
    .encryption-notice {
      margin-top: 16px;
      font-size: var(--text-xs);
      color: var(--rw-mid);
      padding: 8px 12px;
      background: rgba(110, 231, 183, 0.05);
      border-radius: var(--radius-sm);
      border: 1px solid rgba(110, 231, 183, 0.1);
    }

    /* Region Picker */
    .region-picker {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 10px;
    }
    .region-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 12px 14px;
      background: var(--surface-1);
      border: 1px solid rgba(91, 106, 189, 0.2);
      border-radius: var(--radius-md);
      cursor: pointer;
      transition: all var(--transition-fast);
    }
    .region-item.selected {
      border-color: var(--rw-accent);
      background: rgba(119, 126, 240, 0.08);
    }
    .region-checkbox { display: none; }
    .region-icon { font-size: 1.1rem; }
    .region-name {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-white);
      flex: 1;
    }
    .region-cloud {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
    }
    .selected-summary {
      margin-top: 16px;
      font-family: var(--font-body);
      font-size: var(--text-body);
      color: var(--rw-mist);
    }

    /* Review */
    .review-card { margin-bottom: 24px; }
    .review-row {
      display: flex;
      justify-content: space-between;
      padding: 10px 0;
      border-bottom: 1px solid rgba(91, 106, 189, 0.1);
    }
    .review-row:last-child { border-bottom: none; }
    .review-label { color: var(--rw-mid); font-size: var(--text-body); }
    .review-value {
      font-family: var(--font-mono);
      color: var(--rw-white);
      font-weight: 500;
    }
    .deploy-btn {
      width: 100%;
      padding: 14px;
      font-size: var(--text-section);
      justify-content: center;
    }
    .deploy-btn.deploying { opacity: 0.7; pointer-events: none; }
    .spinner {
      display: inline-block;
      width: 16px;
      height: 16px;
      border: 2px solid rgba(255,255,255,0.3);
      border-top-color: white;
      border-radius: 50%;
      animation: rw-spin 0.6s linear infinite;
    }

    .step-nav {
      display: flex;
      justify-content: space-between;
    }

    @media (max-width: 768px) {
      .editor-area { grid-template-columns: 1fr; }
    }
  `]
})
export class OnboardingComponent {
  currentStep = signal(1);
  yamlContent = signal(`name: my-workload\nruntime: go\nreplicas: 3\nport: 8080\nhealth_check: /healthz`);
  activeCloud = signal('AWS');
  selectedRegions = signal<string[]>(['us-east-1', 'eastus']);
  deploying = signal(false);
  deployed = signal(false);

  steps = [
    { number: 1, label: 'Descriptor' },
    { number: 2, label: 'Credentials' },
    { number: 3, label: 'Regions' },
    { number: 4, label: 'Deploy' },
  ];

  cloudTabs = [
    { name: 'AWS', color: '#F59E0B' },
    { name: 'Azure', color: '#3B82F6' },
    { name: 'GCP', color: '#EF4444' },
  ];

  availableRegions = [
    { id: 'us-east-1', name: 'US East (N. Virginia)', cloud: 'AWS', color: '#F59E0B' },
    { id: 'eu-west-1', name: 'EU West (Ireland)', cloud: 'AWS', color: '#F59E0B' },
    { id: 'ap-southeast-1', name: 'Asia Pacific (Singapore)', cloud: 'AWS', color: '#F59E0B' },
    { id: 'eastus', name: 'East US', cloud: 'Azure', color: '#3B82F6' },
    { id: 'westeurope', name: 'West Europe', cloud: 'Azure', color: '#3B82F6' },
    { id: 'us-central1', name: 'US Central', cloud: 'GCP', color: '#EF4444' },
    { id: 'europe-west4', name: 'Europe West', cloud: 'GCP', color: '#EF4444' },
  ];

  toggleRegion(id: string): void {
    this.selectedRegions.update(regions => {
      if (regions.includes(id)) return regions.filter(r => r !== id);
      return [...regions, id];
    });
  }

  parseName(): string {
    const match = this.yamlContent().match(/name:\s*(.+)/);
    return match ? match[1].trim() : 'unnamed';
  }

  uniqueClouds(): string {
    const clouds = new Set(
      this.selectedRegions().map(id => {
        const region = this.availableRegions.find(r => r.id === id);
        return region?.cloud || '';
      })
    );
    return Array.from(clouds).join(', ');
  }

  deploy(): void {
    this.deploying.set(true);
    setTimeout(() => {
      this.deploying.set(false);
      this.deployed.set(true);
    }, 3000);
  }

  getInputValue(event: Event): string {
    return (event.target as HTMLTextAreaElement).value;
  }
}
