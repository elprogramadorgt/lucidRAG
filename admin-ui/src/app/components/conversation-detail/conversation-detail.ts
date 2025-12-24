import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { ConversationService } from '../../services/conversation.service';
import { Conversation, Message } from '../../models/conversation.model';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-conversation-detail',
  imports: [CommonModule, RouterModule],
  templateUrl: './conversation-detail.html',
  styleUrl: './conversation-detail.scss',
})
export class ConversationDetail implements OnInit {
  conversationId = '';
  conversation = signal<Conversation | null>(null);
  messages = signal<Message[]>([]);
  total = signal(0);
  loading = signal(true);
  error = signal<string | null>(null);
  limit = 50;
  offset = 0;

  constructor(
    private route: ActivatedRoute,
    private conversationService: ConversationService,
    public authService: AuthService
  ) {}

  ngOnInit(): void {
    this.conversationId = this.route.snapshot.paramMap.get('id') || '';
    if (this.conversationId) {
      this.loadConversation();
      this.loadMessages();
    }
  }

  loadConversation(): void {
    this.conversationService.getConversation(this.conversationId).subscribe({
      next: (conv) => {
        this.conversation.set(conv);
      },
      error: (err) => {
        console.error('Failed to load conversation', err);
      }
    });
  }

  loadMessages(): void {
    this.loading.set(true);
    this.error.set(null);

    this.conversationService.getMessages(this.conversationId, this.limit, this.offset).subscribe({
      next: (response) => {
        this.messages.set(response.messages || []);
        this.total.set(response.total);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.error?.error || 'Failed to load messages');
        this.loading.set(false);
      }
    });
  }

  nextPage(): void {
    if (this.offset + this.limit < this.total()) {
      this.offset += this.limit;
      this.loadMessages();
    }
  }

  prevPage(): void {
    if (this.offset > 0) {
      this.offset = Math.max(0, this.offset - this.limit);
      this.loadMessages();
    }
  }

  logout(): void {
    this.authService.logout();
  }

  formatTime(dateStr: string): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleTimeString();
  }

  formatDate(dateStr: string): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString();
  }

  get rangeEnd(): number {
    return Math.min(this.offset + this.limit, this.total());
  }
}
