import { Routes } from '@angular/router';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from './core/services/auth.service';

const authGuard = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return auth.isAuthenticated() ? true : router.createUrlTree(['/login']);
};

const publicGuard = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return !auth.isAuthenticated() ? true : router.createUrlTree(['/dashboard']);
};

export const routes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  {
    path: 'login',
    canActivate: [publicGuard],
    loadComponent: () => import('./pages/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'workloads',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/workloads/workloads.component').then(m => m.WorkloadsComponent)
  },
  {
    path: 'builds',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/builds/builds.component').then(m => m.BuildsComponent)
  },
  {
    path: 'cluster',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/cluster/cluster.component').then(m => m.ClusterComponent)
  },
  {
    path: 'raft',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/raft/raft.component').then(m => m.RaftComponent)
  },
  {
    path: 'replication',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/replication/replication.component').then(m => m.ReplicationComponent)
  },
  {
    path: 'failovers',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/failovers/failovers.component').then(m => m.FailoversComponent)
  },
  {
    path: 'onboarding',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/onboarding/onboarding.component').then(m => m.OnboardingComponent)
  },
  {
    path: 'settings',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/settings/settings.component').then(m => m.SettingsComponent)
  },
  { path: '**', redirectTo: 'dashboard' }
];
