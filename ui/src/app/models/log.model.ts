export interface LogEntry {
  id: string;
  level: string;
  message: string;
  timestamp: string;
  source?: string;
  request_id?: string;
  user_id?: string;
  attrs?: Record<string, unknown>;
}

export interface LogFilter {
  level?: string;
  search?: string;
  start_time?: string;
  end_time?: string;
  request_id?: string;
  source?: string;
  limit?: number;
  offset?: number;
}

export interface LogStats {
  total_count: number;
  level_counts: Record<string, number>;
  start_time: string;
  end_time: string;
}

export interface LogResponse {
  logs: LogEntry[];
  total: number;
  limit: number;
  offset: number;
}

export interface ServerInfo {
  status: string;
  environment: string;
  version: string;
  uptime: string;
  uptime_seconds: number;
  started_at: string;
  database: DatabaseStatus;
  runtime: RuntimeInfo;
  endpoints: EndpointInfo[];
}

export interface DatabaseStatus {
  status: string;
  latency?: string;
  latency_ms?: number;
}

export interface RuntimeInfo {
  go_version: string;
  num_cpu: number;
  num_goroutine: number;
  mem_alloc_mb: number;
  mem_sys_mb: number;
}

export interface EndpointInfo {
  path: string;
  method: string;
  description: string;
}
