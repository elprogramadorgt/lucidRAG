import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { ConversationService } from '../../services/conversation.service';
import { ChatSession, Message } from '../../models/conversation.model';

@Component({
  selector: 'app-conversation-detail',
  imports: [CommonModule, RouterModule],
  templateUrl: './conversation-detail.html',
  styleUrl: './conversation-detail.scss',
})
export class ConversationDetail implements OnInit {
  session: ChatSession | null = null;
  messages: Message[] = [];
  loading: boolean = false;
  error: string | null = null;
  sessionId: string = '';

  constructor(
    private route: ActivatedRoute,
    private conversationService: ConversationService
  ) {}

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.sessionId = params['id'];
      if (this.sessionId) {
        this.loadSessionData();
      }
    });
  }

  loadSessionData(): void {
    this.loading = true;
    this.error = null;

    this.conversationService.getSession(this.sessionId).subscribe({
      next: (session) => {
        this.session = session;
        this.loadMessages();
      },
      error: (err) => {
        console.error('Error loading session:', err);
        this.error = 'Failed to load conversation details.';
        this.loading = false;
      }
    });
  }

  loadMessages(): void {
    this.conversationService.getMessages(this.sessionId, 100, 0).subscribe({
      next: (response) => {
        this.messages = response.messages;
        this.loading = false;
      },
      error: (err) => {
        console.error('Error loading messages:', err);
        this.error = 'Failed to load messages.';
        this.loading = false;
      }
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  formatTime(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleTimeString();
  }

  isIncomingMessage(message: Message): boolean {
    // Messages from the user's phone number are incoming (from user to system)
    // Messages to the user's phone number are outgoing (from system to user)
    return message.from === this.session?.user_phone_number;
  }
}
