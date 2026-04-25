import { Component } from '@angular/core';

@Component({
  selector: 'app-onboarding',
  standalone: true,
  template: `
    <div class="max-w-2xl mx-auto">
      <h2 class="text-2xl font-bold mb-6">Deploy Your First Workload</h2>
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-gray-600 mb-4">
          Connect your repository, configure cloud regions, and deploy across
          AWS, Azure, and GCP with automatic failover.
        </p>
        <!-- Onboarding wizard steps will be implemented here -->
      </div>
    </div>
  `,
})
export class OnboardingComponent {}
