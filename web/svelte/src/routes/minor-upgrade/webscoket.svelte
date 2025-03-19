export class WebSocketClient {

}
<script lang="ts">
    let { clusterStatus , formModal , packagesInfo , progressDetails , upgradeConfig } = $props();
    import type { UpgradeProgression } from './ambient';
    import { progressInfo } from './progress.svelte.ts';

    export class WebSocketClient {
        private socket: WebSocket | null = null;
        private readonly url: string;
        private shouldReconnect = true;
        private reconnectInterval = 10000;

        constructor(url: string) {
            this.url = url;
        }

        private connect(): void {
            this.socket = new WebSocket(this.url);

            this.socket.onopen = () => {
                console.log('connected to the server');
                // this.sendMessage({ type: 'ping', data: 'health check beat' });
            };

            this.socket.onmessage = (event) => {
                try {
                    const upgradeProgression: UpgradeProgression = JSON.parse(event.data);
                    // this.handleMessage(event);
                    if (upgradeProgression.stateName == 'upgrade config selection') {
                        formModal = true;
                        console.log(upgradeProgression);
                        if ('k8sPackagesInfo' in upgradeProgression.state) {
                            console.log('packageInfo');
                            packagesInfo = upgradeProgression.state.k8sPackagesInfo;
                        }
                    } else if (upgradeProgression.stateName == 'upgrade progression') {
                        if (
                            'selectedConfig' in upgradeProgression.state &&
                            'nodeStatus' in upgradeProgression.state &&
                            'progressDetails' in upgradeProgression.state
                        ) {
                            upgradeConfig = upgradeProgression.state.selectedConfig;
                            clusterStatus = upgradeProgression.state.nodeStatus;
                            progressDetails = upgradeProgression.state.progressDetails;
                        } else {
                            // TODO:
                            //   sent error reply and expect another payload
                        }

                        for (let progressDetail of progressDetails) {
                            progressInfo.forEach((step) => {
                                if (step.stepCount == progressDetail.stepCount) {
                                    step.log = progressDetail.log;
                                }
                            });
                        }
                    }
                } catch (error) {
                    console.error('Error parsing message:', error);
                }
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

</script>