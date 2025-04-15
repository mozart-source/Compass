export interface Message {
  id?: string;
  text: string;
  sender: 'user' | 'assistant' | 'bot';
  timestamp: Date;
}

export interface Position {
  x: number;
  y: number;
} 