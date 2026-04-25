import { Component } from '@angular/core';

@Component({
  selector: 'rw-settings',
  standalone: true,
  template: `
    <div class="settings">
      <h1 class="page-title">Settings</h1>

      <div class="settings-grid">
        <!-- Workspace Settings -->
        <div class="settings-section rw-card-static">
          <h2 class="section-title">Workspace</h2>
          <div class="form-group">
            <label class="form-label">Workspace Name</label>
            <input class="form-input" type="text" value="Production" />
          </div>
          <div class="form-group">
            <label class="form-label">Workspace ID</label>
            <input class="form-input mono" type="text" value="ws-prod-001" readonly />
          </div>
        </div>

        <!-- Team Members -->
        <div class="settings-section rw-card-static">
          <h2 class="section-title">Team Members</h2>
          <div class="team-list">
            <div class="team-member">
              <div class="team-avatar">P</div>
              <div class="team-info">
                <span class="team-name">Prithviraj</span>
                <span class="team-role">Admin</span>
              </div>
            </div>
            <div class="team-member">
              <div class="team-avatar op">O</div>
              <div class="team-info">
                <span class="team-name">Operator-1</span>
                <span class="team-role">Operator</span>
              </div>
            </div>
            <div class="team-member">
              <div class="team-avatar vi">V</div>
              <div class="team-info">
                <span class="team-name">Viewer-1</span>
                <span class="team-role">Viewer</span>
              </div>
            </div>
          </div>
          <button class="rw-btn-ghost" style="margin-top: 12px;">+ Invite Member</button>
        </div>

        <!-- Cloud Connections -->
        <div class="settings-section rw-card-static">
          <h2 class="section-title">Cloud Connections</h2>
          <div class="connection-list">
            <div class="connection">
              <span class="conn-icon material-symbols-outlined" style="color: #F59E0B">cloud</span>
              <span class="conn-name">AWS</span>
              <span class="conn-status connected">Connected</span>
            </div>
            <div class="connection">
              <span class="conn-icon material-symbols-outlined" style="color: #3B82F6">cloud</span>
              <span class="conn-name">Azure</span>
              <span class="conn-status connected">Connected</span>
            </div>
            <div class="connection">
              <span class="conn-icon material-symbols-outlined" style="color: #EF4444">cloud</span>
              <span class="conn-name">GCP</span>
              <span class="conn-status connected">Connected</span>
            </div>
          </div>
        </div>

        <!-- Danger Zone -->
        <div class="settings-section rw-card-static danger">
          <h2 class="section-title danger-title">Danger Zone</h2>
          <p class="danger-desc">Irreversible actions for this workspace.</p>
          <button class="rw-btn-ghost danger-btn">Delete Workspace</button>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .settings { max-width: 900px; margin: 0 auto; }
    .page-title {
      font-family: var(--font-heading);
      font-size: var(--text-title);
      color: var(--rw-white);
      margin-bottom: 24px;
    }
    .settings-grid { display: flex; flex-direction: column; gap: 20px; }
    .settings-section { padding: 24px; }
    .section-title {
      font-family: var(--font-heading);
      font-size: var(--text-section);
      color: var(--rw-white);
      margin-bottom: 16px;
    }
    .form-group { margin-bottom: 14px; }
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
      max-width: 400px;
      padding: 10px 14px;
      background: var(--surface-1);
      border: 1px solid rgba(91, 106, 189, 0.3);
      border-radius: var(--radius-sm);
      color: var(--rw-white);
      font-family: var(--font-body);
      font-size: var(--text-body);
      outline: none;
    }
    .form-input:focus { border-color: var(--rw-accent); }
    .form-input.mono { font-family: var(--font-mono); color: var(--rw-mid); }

    .team-list { display: flex; flex-direction: column; gap: 10px; }
    .team-member {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 14px;
      background: rgba(66, 68, 127, 0.3);
      border-radius: var(--radius-sm);
    }
    .team-avatar {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: var(--rw-accent);
      color: var(--rw-white);
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--font-heading);
      font-size: 0.8rem;
      font-weight: 600;
    }
    .team-avatar.op { background: var(--healthy); color: #1a1a2e; }
    .team-avatar.vi { background: var(--rw-mid); }
    .team-info { display: flex; flex-direction: column; }
    .team-name {
      font-family: var(--font-body);
      font-size: var(--text-body);
      color: var(--rw-white);
    }
    .team-role {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      color: var(--rw-mid);
      text-transform: uppercase;
    }

    .connection-list { display: flex; flex-direction: column; gap: 8px; }
    .connection {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 14px;
      background: rgba(66, 68, 127, 0.3);
      border-radius: var(--radius-sm);
    }
    .conn-icon { font-size: 1.1rem; }
    .conn-name {
      font-family: var(--font-heading);
      font-size: var(--text-body);
      color: var(--rw-white);
      flex: 1;
    }
    .conn-status {
      font-family: var(--font-mono);
      font-size: var(--text-xs);
      padding: 2px 8px;
      border-radius: 4px;
    }
    .conn-status.connected {
      background: rgba(110, 231, 183, 0.12);
      color: var(--healthy);
    }

    .danger {
      border-color: rgba(248, 113, 113, 0.2);
    }
    .danger-title { color: var(--failed); }
    .danger-desc {
      color: var(--rw-mid);
      font-size: var(--text-body);
      margin-bottom: 14px;
    }
    .danger-btn {
      border-color: rgba(248, 113, 113, 0.3) !important;
      color: var(--failed) !important;
    }
    .danger-btn:hover {
      background: rgba(248, 113, 113, 0.1) !important;
      border-color: var(--failed) !important;
    }
  `]
})
export class SettingsComponent {}
