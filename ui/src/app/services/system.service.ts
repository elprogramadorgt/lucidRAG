import { Injectable, signal } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { LogEntry, LogFilter, LogResponse, LogStats, ServerInfo } from '../models/log.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class SystemService {
  logs = signal<LogEntry[]>([]);
  stats = signal<LogStats | null>(null);
  serverInfo = signal<ServerInfo | null>(null);
  loading = signal(false);
  serverInfoLoading = signal(false);

  constructor(private http: HttpClient) {}

  getServerInfo(): Observable<ServerInfo> {
    this.serverInfoLoading.set(true);
    return this.http.get<ServerInfo>(`${environment.apiUrl}/v1/system/info`).pipe(
      tap({
        next: (info) => {
          this.serverInfo.set(info);
          this.serverInfoLoading.set(false);
        },
        error: () => this.serverInfoLoading.set(false),
      })
    );
  }

  getLogs(filter: LogFilter = {}): Observable<LogResponse> {
    let params = new HttpParams();

    if (filter.level) params = params.set('level', filter.level);
    if (filter.search) params = params.set('search', filter.search);
    if (filter.start_time) params = params.set('start_time', filter.start_time);
    if (filter.end_time) params = params.set('end_time', filter.end_time);
    if (filter.request_id) params = params.set('request_id', filter.request_id);
    if (filter.source) params = params.set('source', filter.source);
    if (filter.limit) params = params.set('limit', filter.limit.toString());
    if (filter.offset) params = params.set('offset', filter.offset.toString());

    this.loading.set(true);
    return this.http.get<LogResponse>(`${environment.apiUrl}/v1/system/logs`, { params }).pipe(
      tap({
        next: (response) => {
          this.logs.set(response.logs || []);
          this.loading.set(false);
        },
        error: () => this.loading.set(false),
      })
    );
  }

  getStats(): Observable<LogStats> {
    return this.http.get<LogStats>(`${environment.apiUrl}/v1/system/logs/stats`).pipe(
      tap((stats) => this.stats.set(stats))
    );
  }

  cleanupLogs(days: number): Observable<{ deleted: number; days: number }> {
    return this.http.delete<{ deleted: number; days: number }>(
      `${environment.apiUrl}/v1/system/logs`,
      { params: new HttpParams().set('days', days.toString()) }
    );
  }
}
