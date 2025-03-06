import { json } from '@sveltejs/kit';

export async function connectionInitialization() {

    const socket = new WebSocket("ws://kapetanios.default.cluster.local/minor-upgrade");

    socket.onopen = async () => {
    //     "current step"
    //     prepare response accordingly
    }

    socket.onmessage = (e) => {
        const message = e.data.split(' ').map(b => String.fromCharCode(parseInt(b, 2))).join('');
        // console.log(message);
        // const message = JSON.parse(e.data);
    }

    socket.onerror = (e) => {

    }

    socket.onclose = async () => {

    }
}