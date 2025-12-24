import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login {
  email = '';
  password = '';
  error = signal<string | null>(null);
  loading = signal(false);
  mode = signal<'login' | 'register'>('login');
  name = '';

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  toggleMode(): void {
    this.mode.update(m => m === 'login' ? 'register' : 'login');
    this.error.set(null);
  }

  onSubmit(): void {
    this.error.set(null);
    this.loading.set(true);

    if (this.mode() === 'login') {
      this.authService.login({ email: this.email, password: this.password }).subscribe({
        next: () => {
          this.loading.set(false);
          this.router.navigate(['/']);
        },
        error: (err) => {
          this.loading.set(false);
          this.error.set(err.error?.error || 'Login failed. Please check your credentials.');
        }
      });
    } else {
      this.authService.register({ email: this.email, password: this.password, name: this.name }).subscribe({
        next: () => {
          this.loading.set(false);
          this.authService.login({ email: this.email, password: this.password }).subscribe({
            next: () => this.router.navigate(['/']),
            error: () => {
              this.mode.set('login');
              this.error.set('Registration successful. Please log in.');
            }
          });
        },
        error: (err) => {
          this.loading.set(false);
          this.error.set(err.error?.error || 'Registration failed. Please try again.');
        }
      });
    }
  }
}
