import {CanActivateFn, Router, Routes} from '@angular/router';
import {AuthComponent} from './components/auth/auth.component';
import {DashboardComponent} from './components/dashboard/dashboard.component';
import {inject} from '@angular/core';
import {AuthService} from './core/services/auth.service';
import {filter, map, take} from 'rxjs/operators';

const createAuthGuard = (requireAuth: boolean): CanActivateFn => () => {
  const authService = inject(AuthService);
  const router      = inject(Router);

  return authService.isLoggedIn$.pipe(
    filter(v => v !== null),
    take(1),
    map(isLoggedIn =>
      isLoggedIn === requireAuth
        ? true
        : router.createUrlTree([requireAuth ? '/auth' : '/dashboard'])
    )
  );
};

export const routes: Routes = [
  { path: '',         redirectTo: 'dashboard', pathMatch: 'full' },
  { path: 'auth',     component: AuthComponent,      canActivate: [createAuthGuard(false)] },
  { path: 'dashboard', component: DashboardComponent, canActivate: [createAuthGuard(true)] },
];
