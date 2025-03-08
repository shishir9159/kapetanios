import { json } from '@sveltejs/kit';

class TestClass {
	prop = $state('test');
}

export let message: string;

export async function connectionInitialization(logMessage: string): Promise<void> {

	const socket = new WebSocket('ws://kapetanios.com/minor-upgrade');
	// const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

	socket.onopen = async () => {
		console.log('connected to the server');
		socket.send('next');
		//     "current step"
		//     prepare response accordingly
	};

	socket.onmessage = (e) => {

		const utf8Array: string = e.data;
		console.log(utf8Array);
		message = utf8Array;
		// const message = JSON.parse(e.data);
	};

	socket.onerror = (e) => {
		console.error('WebSocket error:', e);
	};

	socket.onclose = async () => {
		console.log('WebSocket closed');
	};
}