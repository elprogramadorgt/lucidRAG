import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { ChatSession, SessionListResponse, MessageListResponse } from '../models/conversation.model';

@Injectable({
  providedIn: 'root'
})
export class ConversationService {
  private apiUrl = `${environment.apiUrl}/conversations`;

  constructor(private http: HttpClient) { }

  getSessions(limit: number = 50, offset: number = 0): Observable<SessionListResponse> {
    const params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());
    
    return this.http.get<SessionListResponse>(this.apiUrl, { params });
  }

  getSession(id: string): Observable<ChatSession> {
    const params = new HttpParams().set('id', id);
    return this.http.get<ChatSession>(`${this.apiUrl}/session`, { params });
  }

  getMessages(sessionId: string, limit: number = 100, offset: number = 0): Observable<MessageListResponse> {
    const params = new HttpParams()
      .set('session_id', sessionId)
      .set('limit', limit.toString())
      .set('offset', offset.toString());
    
    return this.http.get<MessageListResponse>(`${this.apiUrl}/messages`, { params });
  }
}
