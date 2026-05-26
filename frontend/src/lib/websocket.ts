type WSCallback = (data: any) => void;

interface WSMessage {
  type: string;
  payload: any;
}

class WebSocketClient {
  private ws: WebSocket | null = null;
  private listeners: Map<string, Set<WSCallback>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private isIntentionalClose = false;

  private getWsUrl(): string {
    if (typeof window === 'undefined') return '';
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL;
    if (wsUrl) return wsUrl;
    // Default: use same host with wss/ws protocol
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}`;
  }

  connect(token: string, endpointId?: string): void {
    if (typeof window === 'undefined') return;
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const baseUrl = this.getWsUrl();
    let wsUrl = `${baseUrl}/ws?token=${token}`;
    if (endpointId) {
      wsUrl += `&endpoint_id=${endpointId}`;
    }

    this.isIntentionalClose = false;

    try {
      this.ws = new WebSocket(wsUrl);
    } catch (err) {
      console.error('WebSocket connection failed:', err);
      return;
    }

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

    this.ws.onerror = () => {
      // Silently handle - onclose will fire after this
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

  off(event: string): void {
    this.listeners.delete(event);
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
