import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthComponent } from '../auth/auth.component';
import { DashboardComponent } from '../dashboard/dashboard.component';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, AuthComponent, DashboardComponent],
  template: `
    @if (authService.isLoggedIn$ | async) {
      <app-dashboard></app-dashboard>
    } @else {
      <app-auth></app-auth>
    }
  `
})
export class HomeComponent {
  authService = inject(AuthService);
}
