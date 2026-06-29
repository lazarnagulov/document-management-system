import { Component, inject, signal } from '@angular/core';
import {CommonModule, NgOptimizedImage} from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { AuthService } from '../../core/services/auth.service';
import { HttpErrorResponse } from '@angular/common/http';
import {AuthRequest} from '../../models/auth.model';

@Component({
  selector: 'app-auth',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, NgOptimizedImage],
  templateUrl: './auth.component.html'
})
export class AuthComponent {
  private readonly fb          = inject(FormBuilder);
  private readonly authService = inject(AuthService);

  isLoginMode = signal(true);
  isLoading   = signal(false);
  generalError = signal<string | null>(null);

  form = this.fb.group({
    email:    ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  get email()    { return this.form.controls.email; }
  get password() { return this.form.controls.password; }

  toggleMode(): void {
    this.isLoginMode.update(v => !v);
    this.generalError.set(null);
    this.form.reset();
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.generalError.set(null);
    this.isLoading.set(true);

    const req = this.form.getRawValue() as AuthRequest;

    if (this.isLoginMode()) {
      this.authService.login(req).subscribe({
        error: (err: HttpErrorResponse) => {
          this.isLoading.set(false);
          this.generalError.set(
            err.status === 401
              ? 'Invalid email or password.'
              : 'Something went wrong. Please try again.'
          );
        },
        complete: () => this.isLoading.set(false),
      });
    } else {
      this.authService.register(req).subscribe({
        next: () => {
          this.authService.login(req).subscribe({
            error: () => {
              this.isLoading.set(false);
              this.isLoginMode.set(true);
            },
            complete: () => this.isLoading.set(false),
          });
        },
        error: (err: HttpErrorResponse) => {
          this.isLoading.set(false);
          this.generalError.set(
            err.status === 409
              ? 'An account with this email already exists.'
              : 'Registration failed. Please try again.'
          );
        },
      });
    }
  }
}
