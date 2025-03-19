class WebSocketClient {
	private socket: WebSocket | null = null;
	private readonly url: string;
	private shouldReconnect = true;
	private reconnectInterval = 10000;

	constructor(url: string) {
		this.url = url;
		this.connect();
	}

	private connect(): void {
		this.socket = new WebSocket(this.url);

		this.socket.onopen = () => {
			console.log('connected to the server');
			// this.sendMessage({ type: 'ping', data: 'health check beat' });
		};

		this.socket.onmessage = (event) => {
			this.handleMessage(event);
		};

		this.socket.onerror = (error) => {
			console.error('WebSocket error:', error);
		};

		this.socket.onclose = (event) => {
			console.warn(`WebSocket closed: ${event.code} - ${event.reason}`);
			if (this.shouldReconnect) {
				setTimeout(() => this.connect(), this.reconnectInterval);
			}
		};
	}

	private handleMessage(event: MessageEvent): void {
		try {
			const message = JSON.parse(event.data);
			console.log('Received message:', message);
		} catch (error) {
			console.error('Error parsing message:', error);
		}
	}

	public sendMessage(message: object): void {
		if (this.socket && this.socket.readyState === WebSocket.OPEN) {
			this.socket.send(JSON.stringify(message));
		} else {
			console.warn('WebSocket is not open. Message not sent.');
		}
	}

	public close(): void {
		this.shouldReconnect = false;
		this.socket?.close();
	}
}
