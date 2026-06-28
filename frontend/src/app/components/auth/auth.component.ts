import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-auth',
  standalone: true,
  imports: [CommonModule],
  templateUrl: "auth.component.html"
})
export class AuthComponent {
  isLoginMode = true;
  private readonly authService = inject(AuthService);

  emailError: string | null = null;
  passwordError: string | null = null;
  generalError: string | null = null;

  toggleMode() {
    this.isLoginMode = !this.isLoginMode;
    this.clearErrors();
  }

  clearErrors() {
    this.emailError = null;
    this.passwordError = null;
    this.generalError = null;
  }

  onSubmit(event: Event) {
    event.preventDefault();
    this.clearErrors();

    // -------------------------------------------------------------
    // TODO: Connect to backend here.
    // This is a mockup to show how the errors will render visually.
    // -------------------------------------------------------------
    const mockBackendFails = false; // Change to true to test the UI error states

    if (mockBackendFails) {
      this.emailError = "Please enter a valid email address.";
      this.passwordError = "Password must be at least 8 characters long.";
      this.generalError = "Authentication failed. Please check your credentials.";
      return;
    }

    // Success path
    this.authService.login();
  }
}
