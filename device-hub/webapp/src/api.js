import ReconnectingWebSocket from 'reconnecting-websocket';

const rws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');

const onUpdateHandlers = {};

export function addOnUpdatedListener(path, handler) {
    let listeners = onUpdateHandlers[path];
    if(!listeners) {
			onUpdateHandlers[path] = [handler];
    } else {
			listeners.push(handler);
    }
}

export function removeOnUpdatedListener(path, handler) {
    let listeners = onUpdateHandlers[path];
    if(listeners) {
			for(let i = 0; i < listeners.length; i++) {
				if (listeners[i] === handler) {
					listeners.splice(i, 1);
				}
			}
    }
}

rws.addEventListener('open', () => {
    console.log("ws open");
});

rws.addEventListener('close', () => {
    console.log("ws close");
});

rws.addEventListener('message', event => {
    console.log("ws event:");
    console.log(event);
		let lines = event.data.split("\n");
    for(let i = 0; i < lines.length; i++) {
			let line = lines[i];
			let msg = JSON.parse(line);
			let handlers = onUpdateHandlers[msg.path];
			if(handlers) {
	    for(let j = 0; j < handlers.length; j++) {
				handlers[j](msg.value);
	    }
	}
    }
});

rws.addEventListener('error', err => {
    console.log("ws error");
    console.log(err);
});

export function getModules(onLoaded) {
	fetch("/api/modules")
			.then(response => response.json())
			.then(data => onLoaded(data.modules))
}

export function setOutput(path, value) {
		let parts = path.split(".");
		let deviceID = parts[1];
		let outputID = parts[2];
    fetch("/api/devices/" + deviceID + "/outputs/" + outputID, {
			method: 'PUT',
			headers: {
					'Content-Type': 'application/json',
			},
			body: JSON.stringify({ value: value }),
    })
}

export function setInput(path, value) {
	let parts = path.split(".");
	let deviceID = parts[1];
	let outputID = parts[2];
	fetch("/api/devices/" + deviceID + "/inputs/" + outputID, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ value: value }),
	})
}
