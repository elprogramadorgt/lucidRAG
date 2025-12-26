import { Component, signal, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SidebarComponent } from '../layout/sidebar/sidebar';
import { ThemeToggleComponent } from '../shared/theme-toggle/theme-toggle';
import { KeyboardShortcutsModalComponent } from '../shared/keyboard-shortcuts-modal/keyboard-shortcuts-modal';
import { MobileGestureService } from '../../services/mobile-gesture.service';
import { SystemService } from '../../services/system.service';
import { LogFilter } from '../../models/log.model';

@Component({
  selector: 'app-system-logs',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    SidebarComponent,
    ThemeToggleComponent,
    KeyboardShortcutsModalComponent,
  ],
  templateUrl: './system-logs.html',
})
export class SystemLogsComponent implements OnInit {
  private mobileGestureService = inject(MobileGestureService);
  private systemService = inject(SystemService);

  sidebarCollapsed = signal(false);
  activeTab = signal<'overview' | 'logs'>('overview');

  // Filter state
  levelFilter = signal<string>('');
  searchFilter = signal<string>('');
  startDate = signal<string>('');
  endDate = signal<string>('');

  // Pagination
  limit = signal(50);
  offset = signal(0);
  total = signal(0);

  levels = ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'];

  get logs() {
    return this.systemService.logs;
  }

  get loading() {
    return this.systemService.loading;
  }

  get stats() {
    return this.systemService.stats;
  }

  get serverInfo() {
    return this.systemService.serverInfo;
  }

  get serverInfoLoading() {
    return this.systemService.serverInfoLoading;
  }

  ngOnInit(): void {
    this.loadServerInfo();
    this.loadLogs();
    this.loadStats();
  }

  loadServerInfo(): void {
    this.systemService.getServerInfo().subscribe();
  }

  openSidebar(): void {
    this.mobileGestureService.openSidebar();
  }

  loadLogs(): void {
    const filter: LogFilter = {
      limit: this.limit(),
      offset: this.offset(),
    };

    if (this.levelFilter()) filter.level = this.levelFilter();
    if (this.searchFilter()) filter.search = this.searchFilter();
    if (this.startDate()) filter.start_time = new Date(this.startDate()).toISOString();
    if (this.endDate()) filter.end_time = new Date(this.endDate()).toISOString();

    this.systemService.getLogs(filter).subscribe({
      next: (response) => {
        this.total.set(response.total);
      },
    });
  }

  loadStats(): void {
    this.systemService.getStats().subscribe();
  }

  applyFilters(): void {
    this.offset.set(0);
    this.loadLogs();
  }

  clearFilters(): void {
    this.levelFilter.set('');
    this.searchFilter.set('');
    this.startDate.set('');
    this.endDate.set('');
    this.offset.set(0);
    this.loadLogs();
  }

  nextPage(): void {
    if (this.offset() + this.limit() < this.total()) {
      this.offset.update((v) => v + this.limit());
      this.loadLogs();
    }
  }

  prevPage(): void {
    if (this.offset() > 0) {
      this.offset.update((v) => Math.max(0, v - this.limit()));
      this.loadLogs();
    }
  }

  getLevelClass(level: string): string {
    const classes: Record<string, string> = {
      TRACE: 'badge-trace',
      DEBUG: 'badge-debug',
      INFO: 'badge-info',
      WARN: 'badge-warn',
      ERROR: 'badge-error',
      CRITICAL: 'badge-critical',
    };
    return classes[level] || 'badge-info';
  }

  formatTime(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  formatAttrs(attrs: Record<string, unknown> | undefined): string {
    if (!attrs || Object.keys(attrs).length === 0) return '';
    return JSON.stringify(attrs, null, 2);
  }

  refresh(): void {
    this.loadServerInfo();
    this.loadLogs();
    this.loadStats();
  }

  setActiveTab(tab: 'overview' | 'logs'): void {
    this.activeTab.set(tab);
  }

  get currentPage(): number {
    return Math.floor(this.offset() / this.limit()) + 1;
  }

  get totalPages(): number {
    return Math.ceil(this.total() / this.limit());
  }

  get showingEnd(): number {
    return Math.min(this.offset() + this.limit(), this.total());
  }
}
