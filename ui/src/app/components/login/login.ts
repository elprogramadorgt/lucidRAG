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
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule, ThemeToggleComponent, LanguageSwitcherComponent, SocialLoginButtonsComponent],
  templateUrl: './login.html',
  styleUrls: ['./login.scss'],
})
export class LoginComponent {
  email = signal('');
  password = signal('');
  isLoading = signal(false);
  errorMessage = signal('');

  constructor(private authService: AuthService, private router: Router) {}

  onSubmit(): void {
    if (!this.email() || !this.password()) {
      this.errorMessage.set('Please enter both email and password');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set('');

    this.authService
      .login({
        email: this.email(),
        password: this.password(),
      })
      .subscribe({
        next: () => {
          this.isLoading.set(false);
          this.router.navigate(['/dashboard']);
        },
        error: (error) => {
          this.isLoading.set(false);
          this.errorMessage.set(error.error?.error || 'Invalid credentials');
        },
      });
  }

  useDemoCredentials(role: 'admin' | 'user'): void {
    const demoCredentials = {
      admin: { email: 'admin@demo.com', password: 'admin123' },
      user: { email: 'user@demo.com', password: 'user123' },
    };

    this.email.set(demoCredentials[role].email);
    this.password.set(demoCredentials[role].password);
  }
}
