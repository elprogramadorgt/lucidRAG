import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ConversationService } from '../../services/conversation.service';
import { ChatSession } from '../../models/conversation.model';

@Component({
  selector: 'app-conversation-list',
  imports: [CommonModule, RouterModule],
  templateUrl: './conversation-list.html',
  styleUrl: './conversation-list.scss',
})
export class ConversationList implements OnInit {
  sessions: ChatSession[] = [];
  loading: boolean = false;
  error: string | null = null;

  constructor(private conversationService: ConversationService) {}

  ngOnInit(): void {
    this.loadSessions();
  }

  loadSessions(): void {
    this.loading = true;
    this.error = null;

    this.conversationService.getSessions(50, 0).subscribe({
      next: (response) => {
        this.sessions = response.sessions;
        this.loading = false;
      },
      error: (err) => {
        console.error('Error loading sessions:', err);
        this.error = 'Failed to load conversations. Please try again.';
        this.loading = false;
      }
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  getStatusClass(isActive: boolean): string {
    return isActive ? 'status-active' : 'status-inactive';
  }
}
