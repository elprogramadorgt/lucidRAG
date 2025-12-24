export interface Conversation {
  id: string;
  phone_number: string;
  contact_name: string;
  last_message_at: string;
  message_count: number;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  whatsapp_msg_id: string;
  direction: 'incoming' | 'outgoing';
  content: string;
  message_type: string;
  rag_answer: string;
  timestamp: string;
  created_at: string;
}

export interface ConversationListResponse {
  conversations: Conversation[];
  total: number;
  limit: number;
  offset: number;
}

export interface MessageListResponse {
  messages: Message[];
  total: number;
  limit: number;
  offset: number;
}
