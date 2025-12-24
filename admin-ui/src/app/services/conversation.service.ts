import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import {
  Conversation,
  ConversationListResponse,
  MessageListResponse
} from '../models/conversation.model';

@Injectable({
  providedIn: 'root'
})
export class ConversationService {
  private apiUrl = `${environment.apiUrl}/conversations`;

  constructor(private http: HttpClient) { }

  getConversations(limit: number = 20, offset: number = 0): Observable<ConversationListResponse> {
    const params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    return this.http.get<ConversationListResponse>(this.apiUrl, { params });
  }

  getConversation(id: string): Observable<Conversation> {
    return this.http.get<Conversation>(`${this.apiUrl}/${id}`);
  }

  getMessages(conversationId: string, limit: number = 50, offset: number = 0): Observable<MessageListResponse> {
    const params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    return this.http.get<MessageListResponse>(`${this.apiUrl}/${conversationId}/messages`, { params });
  }
}
