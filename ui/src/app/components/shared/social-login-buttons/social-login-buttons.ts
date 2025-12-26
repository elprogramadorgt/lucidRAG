import { Component, OnInit, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../environments/environment';

interface OAuthProviders {
  google: boolean;
  facebook: boolean;
  apple: boolean;
}

@Component({
  selector: 'app-social-login-buttons',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './social-login-buttons.html',
})
export class SocialLoginButtonsComponent implements OnInit {
  private http = inject(HttpClient);

  providers = signal<OAuthProviders>({ google: false, facebook: false, apple: false });
  loading = signal(true);

  ngOnInit(): void {
    this.loadProviders();
  }

  private loadProviders(): void {
    this.http.get<OAuthProviders>(`${environment.apiUrl}/v1/auth/oauth/providers`).subscribe({
      next: (providers) => {
        this.providers.set(providers);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      },
    });
  }

  loginWithGoogle(): void {
    window.location.href = `${environment.apiUrl}/v1/auth/oauth/google`;
  }

  loginWithFacebook(): void {
    window.location.href = `${environment.apiUrl}/v1/auth/oauth/facebook`;
  }

  loginWithApple(): void {
    window.location.href = `${environment.apiUrl}/v1/auth/oauth/apple`;
  }

  hasAnyProvider(): boolean {
    const p = this.providers();
    return p.google || p.facebook || p.apple;
  }
}
