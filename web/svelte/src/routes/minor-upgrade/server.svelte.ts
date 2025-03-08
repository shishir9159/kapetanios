import { json } from '@sveltejs/kit';
import { message } from './upgrading/upgrading.svelte.ts';

class TestClass {
	prop = $state('test');
}


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
		logMessage = utf8Array;
		console.log(logMessage);
		// message = utf8Array;
		// const message = JSON.parse(e.data);
	};

	socket.onerror = (e) => {
		console.error('websocket error:', e);
	};

	socket.onclose = async () => {
		console.log('websocket closed');
	};
}