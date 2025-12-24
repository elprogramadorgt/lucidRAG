import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ConversationService } from '../../services/conversation.service';
import { Conversation } from '../../models/conversation.model';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-conversation-list',
  imports: [CommonModule, RouterModule],
  templateUrl: './conversation-list.html',
  styleUrl: './conversation-list.scss',
})
export class ConversationList implements OnInit {
  conversations = signal<Conversation[]>([]);
  total = signal(0);
  loading = signal(true);
  error = signal<string | null>(null);
  limit = 20;
  offset = 0;

  constructor(
    private conversationService: ConversationService,
    public authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadConversations();
  }

  loadConversations(): void {
    this.loading.set(true);
    this.error.set(null);

    this.conversationService.getConversations(this.limit, this.offset).subscribe({
      next: (response) => {
        this.conversations.set(response.conversations || []);
        this.total.set(response.total);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.error?.error || 'Failed to load conversations');
        this.loading.set(false);
      }
    });
  }

  nextPage(): void {
    if (this.offset + this.limit < this.total()) {
      this.offset += this.limit;
      this.loadConversations();
    }
  }

  prevPage(): void {
    if (this.offset > 0) {
      this.offset = Math.max(0, this.offset - this.limit);
      this.loadConversations();
    }
  }

  logout(): void {
    this.authService.logout();
  }

  formatDate(dateStr: string): string {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
  }

  get totalPages(): number {
    return Math.ceil(this.total() / this.limit);
  }

  get currentPage(): number {
    return Math.floor(this.offset / this.limit) + 1;
  }
}
