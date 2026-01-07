import { Injectable, signal, computed, inject } from '@angular/core';
import { Subject } from 'rxjs';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { ConversationService } from './conversation.service';
import { Conversation } from '../models/conversation.model';

export interface SearchFilters {
  query: string;
  status?: 'active' | 'resolved' | 'pending';
  hasUnread?: boolean;
}

@Injectable({ providedIn: 'root' })
export class SearchService {
  private conversationService = inject(ConversationService);
  private searchTerms$ = new Subject<string>();

  searchQuery = signal('');
  filters = signal<SearchFilters>({ query: '' });
  isSearchOpen = signal(false);

  // Computed filtered conversations
  filteredConversations = computed(() => {
    const conversations = this.conversationService.conversations();
    const currentFilters = this.filters();

    if (!conversations) return [];

    return conversations.filter((conv) => {
      // Text search
      if (currentFilters.query) {
        const query = currentFilters.query.toLowerCase();
        const matchesName = conv.contact_name.toLowerCase().includes(query);
        const matchesMessage = conv.last_message?.toLowerCase().includes(query);
        if (!matchesName && !matchesMessage) return false;
      }

      // Status filter
      if (currentFilters.status && conv.status !== currentFilters.status) {
        return false;
      }

      // Unread filter
      if (currentFilters.hasUnread !== undefined) {
        const hasUnread = conv.unread_count > 0;
        if (currentFilters.hasUnread !== hasUnread) return false;
      }

      return true;
    });
  });

  constructor() {
    // Debounced search
    this.searchTerms$.pipe(debounceTime(300), distinctUntilChanged()).subscribe((query) => {
      this.updateFilters({ query });
    });
  }

  search(query: string): void {
    this.searchQuery.set(query);
    this.searchTerms$.next(query);
  }

  updateFilters(updates: Partial<SearchFilters>): void {
    this.filters.update((current) => ({ ...current, ...updates }));
  }

  clearFilters(): void {
    this.filters.set({ query: '' });
    this.searchQuery.set('');
  }

  openSearch(): void {
    this.isSearchOpen.set(true);
  }

  closeSearch(): void {
    this.isSearchOpen.set(false);
  }

  toggleSearch(): void {
    this.isSearchOpen.update((v) => !v);
  }
}
