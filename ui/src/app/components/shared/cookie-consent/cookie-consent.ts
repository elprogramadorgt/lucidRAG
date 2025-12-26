import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { CookieConsentService } from '../../../services/cookie-consent.service';

@Component({
  selector: 'app-cookie-consent',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './cookie-consent.html',
})
export class CookieConsentComponent {
  cookieConsentService = inject(CookieConsentService);

  accept(): void {
    this.cookieConsentService.acceptAll();
  }

  decline(): void {
    this.cookieConsentService.declineOptional();
  }
}
