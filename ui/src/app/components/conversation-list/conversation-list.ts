import { Component, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ConversationService } from '../../services/conversation.service';
import { SearchService } from '../../services/search.service';
import { SearchBarComponent } from '../shared/search-bar/search-bar';
import { Conversation } from '../../models/conversation.model';

@Component({
  selector: 'app-conversation-list',
  standalone: true,
  imports: [CommonModule, SearchBarComponent],
  templateUrl: './conversation-list.html',
  styleUrls: ['./conversation-list.scss'],
})
export class ConversationListComponent implements OnInit {
  focusedIndex = signal(-1);

  // Use filtered conversations from search service

  constructor(
    public conversationService: ConversationService,
    public searchService: SearchService
  ) {}

  displayedConversations = computed(() => {
    return this.searchService.filteredConversations();
  });

  ngOnInit(): void {
    this.conversationService.getConversations().subscribe({
      next: (data) => this.conversationService.conversations.set(data.conversations),
      error: () => this.conversationService.conversations.set([]),
    });
  }

  selectConversation(conv: Conversation): void {
    this.conversationService.selectConversation(conv);
  }

  isSelected(conv: Conversation): boolean {
    return this.conversationService.selectedConversation()?.id === conv.id;
  }

  navigateUp(): void {
    this.focusedIndex.update((i) => Math.max(0, i - 1));
  }

  navigateDown(): void {
    const max = this.displayedConversations().length - 1;
    this.focusedIndex.update((i) => Math.min(max, i + 1));
  }

  selectFocused(): void {
    const conversations = this.displayedConversations();
    const index = this.focusedIndex();
    if (index >= 0 && index < conversations.length) {
      this.selectConversation(conversations[index]);
    }
  }
}
