import { Component, signal, computed, output, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../../services/auth.service';
import { ThemeToggleComponent } from '../../shared/theme-toggle/theme-toggle';
import { LanguageSwitcherComponent } from '../../shared/language-switcher/language-switcher';
import { MobileGestureService } from '../../../services/mobile-gesture.service';

interface NavItem {
  route: string;
  labelKey: string;
  icon: string;
  requiredRole?: 'admin' | 'user';
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, TranslateModule, ThemeToggleComponent, LanguageSwitcherComponent],
  templateUrl: './sidebar.html',
  host: {
    class: 'contents',
  },
})
export class SidebarComponent {
  private mobileGestureService = inject(MobileGestureService);

  isCollapsed = signal(false);

  // Sync with mobile gesture service
  get isMobileOpen() {
    return this.mobileGestureService.sidebarOpen;
  }

  collapseChange = output<boolean>();
  mobileClose = output<void>();

  private allNavItems: NavItem[] = [
    { route: '/dashboard', labelKey: 'nav.dashboard', icon: 'dashboard' },
    { route: '/conversations', labelKey: 'nav.conversations', icon: 'chat' },
    { route: '/documents', labelKey: 'nav.documents', icon: 'document' },
    { route: '/system', labelKey: 'nav.system', icon: 'system', requiredRole: 'admin' },
  ];

  authService = inject(AuthService);

  // Computed signal that reacts to currentUser changes
  navItems = computed(() => {
    // Access currentUser signal to establish dependency
    const user = this.authService.currentUser();
    return this.allNavItems.filter(item => {
      if (!item.requiredRole) return true;
      return this.authService.hasPermission(item.requiredRole);
    });
  });

  toggleCollapse(): void {
    this.isCollapsed.update((v) => !v);
    this.collapseChange.emit(this.isCollapsed());
  }

  openMobile(): void {
    this.mobileGestureService.openSidebar();
  }

  closeMobile(): void {
    this.mobileGestureService.closeSidebar();
    this.mobileClose.emit();
  }

  logout(): void {
    this.authService.logout();
  }
}
