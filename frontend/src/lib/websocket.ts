type WSCallback = (data: any) => void;

interface WSMessage {
  type: string;
  payload: any;
}

class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private listeners: Map<string, Set<WSCallback>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private isIntentionalClose = false;

  constructor() {
    this.url = process.env.NEXT_PUBLIC_WS_URL || 'wss://ws.webhook.inst.lk';
  }

  connect(token: string, endpointId?: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    let wsUrl = `${this.url}/ws?token=${token}`;
    if (endpointId) {
      wsUrl += `&endpoint_id=${endpointId}`;
    }

    this.isIntentionalClose = false;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.emit('connected', {});
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data);
        this.emit(message.type, message.payload);
        this.emit('message', message);
      } catch (error) {
        console.error('WebSocket parse error:', error);
      }
    };

    this.ws.onclose = () => {
      this.emit('disconnected', {});
      if (!this.isIntentionalClose && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnect(token, endpointId);
      }
    };

    this.ws.onerror = (error) => {
      this.emit('error', error);
    };
  }

  private reconnect(token: string, endpointId?: string): void {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      this.connect(token, endpointId);
    }, Math.min(delay, 30000));
  }

  disconnect(): void {
    this.isIntentionalClose = true;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  on(event: string, callback: WSCallback): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);

    // Return unsubscribe function
    return () => {
      this.listeners.get(event)?.delete(callback);
    };
  }

  private emit(event: string, data: any): void {
    this.listeners.get(event)?.forEach((callback) => {
      try {
        callback(data);
      } catch (error) {
        console.error('WebSocket listener error:', error);
      }
    });
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const wsClient = new WebSocketClient();
