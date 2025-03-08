import { json } from '@sveltejs/kit';

export async function connectionInitialization(logMessage: string): Promise<void> {
    let message: string = $state('');

	const socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
	// const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

	socket.onopen = async () => {
		console.log('connected to the server');
		socket.send('next');
		//     "current step"
		//     prepare response accordingly
	};

	socket.onmessage = (e) => {
		// const utf8Array = e.data.split(' ');

		const utf8Array = e.data;
		console.log(utf8Array);
		const decoder = new TextDecoder('utf-8');

		const str = decoder.decode(utf8Array);
		console.log('str: ', str);

		logMessage = str;
		// const message = e.data.split(' ').map((b:Uint8Array) => String.fromCharCode(parseInt(b, 2))).join('');
		// console.log(message);
		// const message = JSON.parse(e.data);
	};

	socket.onerror = (e) => {
		console.error('WebSocket error:', e);
	};

	socket.onclose = async () => {
		console.log('WebSocket closed');
	};
}