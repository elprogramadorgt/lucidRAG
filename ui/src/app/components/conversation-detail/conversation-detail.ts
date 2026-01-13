import { Component, signal, ViewChild, ElementRef, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ConversationService } from '../../services/conversation.service';
import { OnlineStatusIndicatorComponent } from '../shared/online-status-indicator/online-status-indicator';

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
  isTyping = signal(false);

  constructor(public conversationService: ConversationService) {}

  // displayedConversations = computed(() => {
  //   return this.conversationService.getConversationMessages();
  // });

  displayedMessages = computed(() => this.conversationService.messages());

  sendMessage(): void {
    const message = this.replyMessage().trim();
    if (!message) return;

    // TODO: Send message via API
    console.log('Sending message:', message);
    this.replyMessage.set('');
  }

  onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  focusInput(): void {
    this.messageInput?.nativeElement?.focus();
  }

  // Simulate online status based on last message time
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
