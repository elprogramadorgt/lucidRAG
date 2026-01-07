import { Injectable, signal } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import {
  Conversation,
  ConversationsResponse,
  Message,
  MessagesResponse,
  SendReplyRequest,
  SendReplyResponse,
  ToggleBotRequest,
  ToggleBotResponse,
  Platform,
  ConversationStatus,
} from '../models/conversation.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class ConversationService {
  conversations = signal<Conversation[]>([]);
  selectedConversation = signal<Conversation | null>(null);
  messages = signal<Message[]>([]);

  constructor(private http: HttpClient) {}

  getConversations(
    platform?: Platform,
    status?: ConversationStatus,
    page: number = 1,
    limit: number = 20
  ): Observable<ConversationsResponse> {
    let params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    if (platform) {
      params = params.set('platform', platform);
    }
    if (status) {
      params = params.set('status', status);
    }

    return this.http
      .get<ConversationsResponse>(`${environment.apiUrl}/v1/conversations`, { params })
      .pipe(
        tap((response) => {
          this.conversations.set(response.conversations);
        })
      );
  }

  getConversationMessages(
    conversationId: string,
    limit: number = 50
  ): Observable<MessagesResponse> {
    const params = new HttpParams().set('limit', limit.toString());

    return this.http
      .get<MessagesResponse>(`${environment.apiUrl}/v1/conversations/${conversationId}/messages`, {
        params,
      })
      .pipe(
        tap((response) => {
          this.messages.set(response.messages);
        })
      );
  }

  sendReply(conversation_id: string, request: SendReplyRequest): Observable<SendReplyResponse> {
    return this.http
      .post<SendReplyResponse>(
        `${environment.apiUrl}/v1/conversations/${conversation_id}/reply`,
        request
      )
      .pipe(
        tap((response) => {
          const newMessage: Message = {
            id: response.message_id,
            conversation_id: conversation_id,
            sender: 'agent',
            text: request.text,
            timestamp: response.timestamp,
            agent_name: 'You',
          };
          this.messages.update((msgs) => [...msgs, newMessage]);
        })
      );
  }

  toggleBot(conversationId: string, enabled: boolean): Observable<ToggleBotResponse> {
    const request: ToggleBotRequest = { enabled };

    return this.http
      .post<ToggleBotResponse>(
        `${environment.apiUrl}/v1/conversations/${conversationId}/toggle-bot`,
        request
      )
      .pipe(
        tap((response) => {
          this.conversations.update((convs) =>
            convs.map((conv) =>
              conv.id === conversationId ? { ...conv, bot_enabled: response.bot_enabled } : conv
            )
          );

          const selected = this.selectedConversation();
          if (selected && selected.id === conversationId) {
            this.selectedConversation.set({ ...selected, bot_enabled: response.bot_enabled });
          }
        })
      );
  }

  selectConversation(conversation: Conversation): void {
    this.selectedConversation.set(conversation);
  }

  clearMessages(): void {
    this.messages.set([]);
  }
}
