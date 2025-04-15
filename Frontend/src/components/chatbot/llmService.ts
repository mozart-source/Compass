import axios from 'axios';
import { getApiUrls } from '@/config';

// Get API URLs using the configuration system
const { PYTHON_API_URL } = getApiUrls();

// Types
export interface LLMRequest {
  prompt: string;
  context?: Record<string, any>;
  domain?: string;
  model_parameters?: Record<string, any>;
  previous_messages?: Array<{sender: string; text: string}>;
  session_id?: string;
}

export interface LLMResponse {
  response: string;
  intent?: string;
  target?: string;
  description?: string;
  rag_used: boolean;
  cached: boolean;
  confidence: number;
  error?: boolean;
  error_message?: string;
  session_id?: string;
  tool_used?: string;
  tool_args?: Record<string, any>;
  tool_success?: boolean;
}

// Local storage key for session ID
const SESSION_ID_KEY = 'ai_conversation_session_id';

// Helper to get or create a session ID
const getOrCreateSessionId = (): string => {
  const existingSessionId = localStorage.getItem(SESSION_ID_KEY);
  if (existingSessionId) {
    return existingSessionId;
  }
  
  // Create a new UUID for the session
  const newSessionId = crypto.randomUUID();
  localStorage.setItem(SESSION_ID_KEY, newSessionId);
  return newSessionId;
};

// LLM Service
export const llmService = {
  // Get current session ID
  getSessionId: (): string => {
    return getOrCreateSessionId();
  },
  
  // Create a new session ID (useful for starting a new conversation)
  createNewSession: (): string => {
    const newSessionId = crypto.randomUUID();
    localStorage.setItem(SESSION_ID_KEY, newSessionId);
    return newSessionId;
  },
  
  // Clear the current session
  clearSession: async (): Promise<void> => {
    const token = localStorage.getItem('token');
    if (!token) throw new Error('Authentication required');
    
    const sessionId = getOrCreateSessionId();
    
    try {
      await axios.post(
        `${PYTHON_API_URL}/ai/clear-session`,
        { session_id: sessionId },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      // Create a new session after clearing
      llmService.createNewSession();
    } catch (error) {
      console.error('Failed to clear session:', error);
      throw error;
    }
  },

  // Stream response from the LLM
  streamResponse: async function* (prompt: string, previousMessages?: Array<{sender: string; text: string}>) {
    const token = localStorage.getItem('token');
    if (!token) {
      console.error("No authentication token found in localStorage");
      yield JSON.stringify({ error: "Authentication required. Please log in again." });
      return;
    }
    
    // Get the current session ID
    const sessionId = getOrCreateSessionId();

    try {
      console.log('Connecting to SSE endpoint:', `${PYTHON_API_URL}/ai/process/stream`);
      console.log('With session ID:', sessionId);
      console.log('Auth token available:', !!token);
      console.log('Auth token preview:', token.substring(0, 10) + '...');
      
      // Set up fetch with proper headers for SSE
      const response = await fetch(`${PYTHON_API_URL}/ai/process/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
          'Accept': 'text/event-stream',
        },
        body: JSON.stringify({ 
          prompt,
          previous_messages: previousMessages,
          session_id: sessionId
        }),
      });

      console.log('SSE response status:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('LLM SSE Error:', errorText);
        yield JSON.stringify({ error: `Stream request failed: ${response.status} ${response.statusText}` });
        return;
      }

      // Get response reader
      const reader = response.body?.getReader();
      if (!reader) {
        console.error('Response body is not readable');
        yield JSON.stringify({ error: 'Response body is not readable' });
        return;
      }

      console.log('SSE connection established, processing stream...');
      const decoder = new TextDecoder();
      
      // Keep track of tool info
      let toolInfo = {
        name: null,
        success: null
      };
      
      // Process the stream
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          console.log('SSE stream complete');
          break;
        }
        
        const chunk = decoder.decode(value, { stream: true });
        console.log('Received chunk:', chunk.substring(0, 50) + (chunk.length > 50 ? '...' : ''));
        
        const lines = chunk.split('\n\n');
        console.log(`Chunk contains ${lines.length} events`);
        
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.substring(6).trim();
            
            if (data === '[DONE]') {
              console.log('Stream complete event received');
              break;
            }
            
            try {
              // Parse the JSON data
              const parsed = JSON.parse(data);
              console.log('Parsed SSE data:', parsed);
              
              // Check for errors
              if (parsed.error) {
                console.error('Stream error:', parsed.error);
                yield JSON.stringify({ error: parsed.error });
                return;
              }
              
              // Save tool info if present
              if (parsed.tool_used) {
                toolInfo.name = parsed.tool_used;
                toolInfo.success = parsed.tool_success;
                console.log('Tool used:', toolInfo);
              }
              
              // Check for completion - don't yield this
              if (parsed.complete) {
                console.log('Completion signal received');
                if (toolInfo.name) {
                  console.log(`Completed processing with tool: ${toolInfo.name}, success: ${toolInfo.success}`);
                }
                continue;
              }
              
              // Yield the token
              if (parsed.token) {
                console.log('Yielding token:', parsed.token.substring(0, 20) + (parsed.token.length > 20 ? '...' : ''));
                yield parsed.token;
              }
            } catch (e) {
              console.warn('Error parsing SSE data:', e, 'Raw data:', data);
              // Handle non-JSON data
              if (data && data !== '[DONE]') {
                console.log('Yielding non-JSON data');
                yield data;
              }
            }
          }
        }
      }
    } catch (error) {
      console.error('Error in stream response:', error);
      yield JSON.stringify({ error: error instanceof Error ? error.message : 'Unknown error' });
    }
  }
};

// Custom hook for streaming responses
export const useStreamingLLMResponse = () => {
  return {
    streamResponse: (prompt: string, previousMessages?: Array<{sender: string; text: string}>) => {
      return llmService.streamResponse(prompt, previousMessages);
    },
  };
};

// Hook to manage conversation sessions
export const useConversationSession = () => {
  return {
    getSessionId: llmService.getSessionId,
    createNewSession: llmService.createNewSession,
    clearSession: llmService.clearSession
  };
};

export default llmService;
