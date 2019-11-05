import ReconnectingWebSocket from 'reconnecting-websocket';

const rws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');
//const rws = new ReconnectingWebSocket('ws://localhost:8085/ws');

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

const onModuleChangedHandlers = [];

export function addOnModuleChangedListener(handler) {
		onModuleChangedHandlers.push(handler);
}

export function removeOnModuleChangedListener(handler) {
		for(let i = 0; i < onModuleChangedHandlers.length; i++) {
			if (onModuleChangedHandlers[i] === handler) {
				onModuleChangedHandlers.splice(i, 1);
			}
		}
}

rws.addEventListener('message', event => {
		let lines = event.data.split("\n");
    for(let i = 0; i < lines.length; i++) {
			let line = lines[i];
			let msg = JSON.parse(line);
			switch(msg.type) {
				case "value":
					let handlers = onUpdateHandlers[msg.path];
					if(handlers) {
						for (let j = 0; j < handlers.length; j++) {
							handlers[j](msg.value);
						}
					}
					break;
				case "module-created":
				case "module-deleted":
				case "module-updated":
					for (let j = 0; j < onModuleChangedHandlers.length; j++) {
						onModuleChangedHandlers[j](msg);
					}
					break;
				default:
					console.log("unrecognized event type: " + msg.type);
			}
		}
});

rws.addEventListener('error', err => {
    console.log(err);
});

export function getDataset(onLoaded) {
	fetch("/api/")
			.then(response => response.json())
			.then(data => onLoaded(data))
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
