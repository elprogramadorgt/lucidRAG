import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { WatermarkComponent } from './components/shared/watermark/watermark';
import { CookieConsentComponent } from './components/shared/cookie-consent/cookie-consent';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, WatermarkComponent, CookieConsentComponent],
  template: `
    <router-outlet></router-outlet>
    <app-watermark></app-watermark>
    <app-cookie-consent></app-cookie-consent>
  `,
})
export class App {}
