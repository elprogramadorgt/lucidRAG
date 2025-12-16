export interface Message {
  id: string;
  from: string;
  to: string;
  content: string;
  message_type: string;
  timestamp: string;
  status: string;
}

export interface ChatSession {
  id: string;
  user_phone_number: string;
  started_at: string;
  last_message_at: string;
  is_active: boolean;
  context: string;
}

export interface SessionListResponse {
  sessions: ChatSession[];
  limit: number;
  offset: number;
}

export interface MessageListResponse {
  messages: Message[];
  session_id: string;
  limit: number;
  offset: number;
}
