export type DocumentStatus = 'UPLOADED' | 'PROCESSING' | 'READY' | 'FAILED';

export interface DocumentResponse {
  id: string;
  name: string;
  content_type: string;
  size: number;
  status: DocumentStatus;
  created_at: string;
  updated_at: string;
}

export interface ChatMessage {
  sender: 'user' | 'ai';
  text: string;
  loading?: boolean;
}

export interface UpdateDocumentRequest {
  name: string;
}

export interface QueryRequest {
  question: string;
}

export interface QueryResponse {
  question: string;
  answer: string;
}
