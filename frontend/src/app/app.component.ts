import {Component, inject, OnInit} from '@angular/core';
import { HomeComponent } from './components/home/home.component';
import {AuthService} from './core/services/auth.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [HomeComponent],
  template: `<app-home></app-home>`
})
export class AppComponent implements OnInit {
  title = 'document-management-system';

  private readonly authService = inject(AuthService);

  ngOnInit(): void {
    this.authService.checkSession().subscribe();
  }
}
