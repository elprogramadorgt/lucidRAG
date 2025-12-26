import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { LanguageService, Language } from '../../../services/language.service';

@Component({
  selector: 'app-language-switcher',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './language-switcher.html',
})
export class LanguageSwitcherComponent {
  languageService = inject(LanguageService);
  isOpen = signal(false);

  toggle(): void {
    this.isOpen.update(v => !v);
  }

  close(): void {
    this.isOpen.set(false);
  }

  selectLanguage(lang: Language): void {
    this.languageService.setLanguage(lang.code);
    this.close();
  }
}
