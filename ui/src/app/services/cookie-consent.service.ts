import { Injectable, signal } from '@angular/core';

export type ConsentStatus = 'pending' | 'accepted' | 'declined';

@Injectable({
  providedIn: 'root',
})
export class CookieConsentService {
  private readonly CONSENT_KEY = 'lucidrag_cookie_consent';

  consentStatus = signal<ConsentStatus>('pending');
  showBanner = signal<boolean>(false);

  constructor() {
    this.loadConsent();
  }

  acceptAll(): void {
    this.setConsent('accepted');
  }

  declineOptional(): void {
    // Auth cookies are essential and always allowed
    // This declines only optional/analytics cookies
    this.setConsent('declined');
  }

  private setConsent(status: ConsentStatus): void {
    localStorage.setItem(this.CONSENT_KEY, status);
    this.consentStatus.set(status);
    this.showBanner.set(false);
  }

  private loadConsent(): void {
    const stored = localStorage.getItem(this.CONSENT_KEY) as ConsentStatus | null;
    if (stored === 'accepted' || stored === 'declined') {
      this.consentStatus.set(stored);
      this.showBanner.set(false);
    } else {
      this.showBanner.set(true);
    }
  }
}
