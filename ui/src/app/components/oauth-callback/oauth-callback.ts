import { Component, OnInit, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, ActivatedRoute } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-oauth-callback',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './oauth-callback.html',
})
export class OAuthCallbackComponent implements OnInit {
  private router = inject(Router);
  private route = inject(ActivatedRoute);
  private authService = inject(AuthService);

  loading = signal(true);
  error = signal<string | null>(null);

  ngOnInit(): void {
    this.handleCallback();
  }

  private handleCallback(): void {
    const success = this.route.snapshot.queryParamMap.get('success');
    const errorMsg = this.route.snapshot.queryParamMap.get('error');

    if (errorMsg) {
      this.error.set(decodeURIComponent(errorMsg));
      this.loading.set(false);
      return;
    }

    if (success === 'true') {
      // Cookie was set by backend, validate session
      this.authService.validateSession().subscribe({
        next: (valid) => {
          this.loading.set(false);
          if (valid) {
            this.router.navigate(['/dashboard']);
          } else {
            this.error.set('Failed to validate session');
          }
        },
        error: () => {
          this.loading.set(false);
          this.error.set('Failed to validate session');
        },
      });
    } else {
      this.loading.set(false);
      this.error.set('OAuth authentication failed');
    }
  }

  goToLogin(): void {
    this.router.navigate(['/login']);
  }
}
