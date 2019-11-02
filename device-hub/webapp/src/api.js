import ReconnectingWebSocket from 'reconnecting-websocket';

const rws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');

const onUpdateHandlers = {};

export function addOnUpdatedListener(path, handler) {
    var listeners = onUpdateHandlers[path];
    if(!listeners) {
	onUpdateHandlers[path] = [handler];
    } else {
	listeners.push(handler);
    }
}

export function removeOnUpdatedListener(path, handler) {
    var listeners = onUpdateHandlers[path];
    if(listeners) {
	for(var i = 0; i < listeners.length; i++) {
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
    var lines = event.data.split("\n");
    for(var i = 0; i < lines.length; i++) {
	var line = lines[i];
	var msg = JSON.parse(line);
	var handlers = onUpdateHandlers[msg.path];
	if(handlers) {
	    for(var j = 0; j < handlers.length; j++) {
		handlers[j](msg.value);
	    }
	}
    }
});

rws.addEventListener('error', err => {
    console.log("ws error");
    console.log(err);
});

export function setOutput(deviceID, outputID, value) {
    fetch("/api/devices/" + deviceID + "/outputs/" + outputID, {
	method: 'PUT',
	headers: {
	    'Content-Type': 'application/json',
	},
	body: JSON.stringify({ value: value }),
    })
}

export function setInput(deviceID, outputID, value) {
	fetch("/api/devices/" + deviceID + "/inputs/" + outputID, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ value: value }),
	})
}
