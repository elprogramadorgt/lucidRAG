export type Platform = 'whatsapp' | 'instagram' | 'telegram' | 'facebook';
export type ConversationStatus = 'active' | 'archived' | 'resolved';

export interface Conversation {
  id: string;
  platform: Platform;
  contact_name: string;
  contact_id: string;
  contact_avatar?: string;
  last_message: string;
  last_message_time: string;
  content: string;
  unread_count: number;
  bot_enabled: boolean;
  average_response_time: number;
  total_messages: number;
  status: ConversationStatus;
}

export interface ConversationsResponse {
  conversations: Conversation[];
  total: number;
  page: number;
  totalPages: number;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender: 'user' | 'bot' | 'agent';
  content: string;
  timestamp: string;
  response_time?: number;
  agent_name?: string;
}

export interface MessagesResponse {
  messages: Message[];
}

export interface SendReplyRequest {
  text: string;
  agent_id: string;
}

export interface SendReplyResponse {
  message_id: string;
  sent: boolean;
  timestamp: string;
}

export interface ToggleBotRequest {
  enabled: boolean;
}

export interface ToggleBotResponse {
  conversation_id: string;
  bot_enabled: boolean;
  message: string;
}
