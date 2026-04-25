import { Component, input } from '@angular/core';

@Component({
  selector: 'rw-skeleton',
  standalone: true,
  template: `
    <div class="skeleton" [style.width]="width()" [style.height]="height()" [style.border-radius]="radius()"></div>
  `,
  styles: [`
    .skeleton {
      background: linear-gradient(
        90deg,
        rgba(81, 82, 162, 0.3) 25%,
        rgba(91, 106, 189, 0.4) 50%,
        rgba(81, 82, 162, 0.3) 75%
      );
      background-size: 200% 100%;
      animation: rw-shimmer 1.5s ease-in-out infinite;
    }
  `]
})
export class SkeletonLoaderComponent {
  width = input<string>('100%');
  height = input<string>('16px');
  radius = input<string>('6px');
}
