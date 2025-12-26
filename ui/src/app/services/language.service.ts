import { Injectable, signal } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

export interface Language {
  code: string;
  name: string;
  flag: string;
}

@Injectable({
  providedIn: 'root',
})
export class LanguageService {
  private readonly LANGUAGE_KEY = 'lucidrag_language';

  readonly languages: Language[] = [
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'es', name: 'Español', flag: '🇪🇸' },
    { code: 'fr', name: 'Français', flag: '🇫🇷' },
    { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
    { code: 'pt', name: 'Português', flag: '🇧🇷' },
    { code: 'zh', name: '中文', flag: '🇨🇳' },
  ];

  currentLanguage = signal<Language>(this.languages[0]);

  constructor(private translate: TranslateService) {
    this.initLanguage();
  }

  private initLanguage(): void {
    // Set available languages
    this.translate.addLangs(this.languages.map(l => l.code));

    // Get stored language or detect from browser
    const storedLang = localStorage.getItem(this.LANGUAGE_KEY);
    const browserLang = this.translate.getBrowserLang();

    let langCode = storedLang || browserLang || 'en';

    // Ensure the language is supported
    if (!this.languages.find(l => l.code === langCode)) {
      langCode = 'en';
    }

    this.setLanguage(langCode);
  }

  setLanguage(code: string): void {
    const language = this.languages.find(l => l.code === code);
    if (language) {
      this.translate.use(code);
      this.currentLanguage.set(language);
      localStorage.setItem(this.LANGUAGE_KEY, code);
      document.documentElement.lang = code;
    }
  }

  getLanguageByCode(code: string): Language | undefined {
    return this.languages.find(l => l.code === code);
  }
}
