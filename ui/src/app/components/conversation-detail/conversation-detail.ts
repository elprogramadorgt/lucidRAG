import { Component, signal, ViewChild, ElementRef, computed, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ConversationService } from '../../services/conversation.service';
import { OnlineStatusIndicatorComponent } from '../shared/online-status-indicator/online-status-indicator';
import { firstValueFrom } from 'rxjs';
import { SendReplyRequest } from '../../models/conversation.model';

@Component({
  selector: 'app-conversation-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, OnlineStatusIndicatorComponent],
  templateUrl: './conversation-detail.html',
  styleUrls: ['./conversation-detail.scss'],
})
export class ConversationDetailComponent {
  @ViewChild('messageInput') messageInput!: ElementRef<HTMLInputElement>;
  @ViewChild('messagesContainer') messagesContainer!: ElementRef<HTMLDivElement>;

  replyMessage = signal('');
  sending = signal(false);

  constructor(public conversationService: ConversationService) {
    // Auto-scroll when messages change
    effect(() => {
      this.displayedMessages();
      queueMicrotask(() => this.scrollToBottom());
    });
  }

  displayedMessages = computed(() => this.conversationService.messages());

  async sendMessage(): Promise<void> {
    const text = this.replyMessage().trim();
    const conv = this.conversationService.selectedConversation();

    if (!conv || !text || this.sending()) return;

    this.sending.set(true);

    // clear input immediately (UX)
    this.replyMessage.set('');
    this.focusInput();

    const req: SendReplyRequest = { text, agent_id: 'agent-123' }; // TODO: replace with real agent ID

    try {
      // IMPORTANT: service already appends the message in tap()
      await firstValueFrom(this.conversationService.sendReply(conv.id, req));
      this.scrollToBottom();
    } catch (err) {
      console.error('Failed to send message:', err);

      // restore text so user can retry
      this.replyMessage.set(text);
      this.focusInput();
    } finally {
      this.sending.set(false);
    }
  }

  onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void this.sendMessage();
    }
  }

  focusInput(): void {
    this.messageInput?.nativeElement?.focus();
  }

  scrollToBottom(): void {
    const el = this.messagesContainer?.nativeElement;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }

  getOnlineStatus(): 'online' | 'away' | 'offline' {
    const conv = this.conversationService.selectedConversation();
    if (!conv) return 'offline';

    const lastMessageTime = new Date(conv.last_message_time);
    const now = new Date();
    const diffMins = (now.getTime() - lastMessageTime.getTime()) / 60000;

    if (diffMins < 5) return 'online';
    if (diffMins < 30) return 'away';
    return 'offline';
  }
}
