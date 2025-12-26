import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../services/auth.service';
import { ThemeToggleComponent } from '../shared/theme-toggle/theme-toggle';
import { LanguageSwitcherComponent } from '../shared/language-switcher/language-switcher';
import { SocialLoginButtonsComponent } from '../shared/social-login-buttons/social-login-buttons';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule, ThemeToggleComponent, LanguageSwitcherComponent, SocialLoginButtonsComponent],
  templateUrl: './register.html',
  styleUrls: ['./register.scss'],
})
export class RegisterComponent {
  firstName = signal('');
  lastName = signal('');
  email = signal('');
  password = signal('');
  confirmPassword = signal('');
  isLoading = signal(false);
  errorMessage = signal('');

  constructor(private authService: AuthService, private router: Router) {}

  onSubmit(): void {
    if (!this.firstName() || !this.lastName() || !this.email() || !this.password()) {
      this.errorMessage.set('Please fill in all fields');
      return;
    }

    if (this.password().length < 8) {
      this.errorMessage.set('Password must be at least 8 characters');
      return;
    }

    if (this.password() !== this.confirmPassword()) {
      this.errorMessage.set('Passwords do not match');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set('');

    this.authService
      .register({
        email: this.email(),
        password: this.password(),
        first_name: this.firstName(),
        last_name: this.lastName(),
      })
      .subscribe({
        next: () => {
          this.isLoading.set(false);
          this.router.navigate(['/dashboard']);
        },
        error: (error) => {
          this.isLoading.set(false);
          this.errorMessage.set(error.error?.error || 'Registration failed');
        },
      });
  }
}
