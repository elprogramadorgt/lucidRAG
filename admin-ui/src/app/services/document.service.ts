import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { Document, RAGQuery, RAGResponse } from '../models/document.model';

@Injectable({
  providedIn: 'root'
})
export class DocumentService {
  private apiUrl = `${environment.apiUrl}/documents`;
  private ragUrl = `${environment.apiUrl}/rag`;

  constructor(private http: HttpClient) { }

  getDocuments(limit: number = 10, offset: number = 0): Observable<{ documents: Document[], limit: number, offset: number }> {
    const params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());
    
    return this.http.get<{ documents: Document[], limit: number, offset: number }>(this.apiUrl, { params });
  }

  getDocument(id: string): Observable<Document> {
    const params = new HttpParams().set('id', id);
    return this.http.get<Document>(this.apiUrl, { params });
  }

  createDocument(document: Partial<Document>): Observable<{ id: string, message: string }> {
    return this.http.post<{ id: string, message: string }>(this.apiUrl, document);
  }

  updateDocument(document: Document): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(this.apiUrl, document);
  }

  deleteDocument(id: string): Observable<{ message: string }> {
    const params = new HttpParams().set('id', id);
    return this.http.delete<{ message: string }>(this.apiUrl, { params });
  }

  queryRAG(query: RAGQuery): Observable<RAGResponse> {
    return this.http.post<RAGResponse>(`${this.ragUrl}/query`, query);
  }
}
