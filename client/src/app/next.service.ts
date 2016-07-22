import {Injectable} from '@angular/core';
import {Subject} from 'rxjs/Subject';

export class StartMessage {
	constructor(
		public peerName: string
	) {};
}

export class EndMessage {
	constructor() {};
}

@Injectable()
export class NextService {
	private websocket: WebSocket;
	private msgSubject = new Subject<StartMessage | EndMessage>();
	msgObservable = this.msgSubject.asObservable();

	private matchID: number = null;
	private peerName: string;

	private localStream: MediaStream = null;

	getUserMedia(): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			navigator.getUserMedia(
				{
					audio: true,
					video: true,
				},
				localMediaStream => {
					this.localStream = localMediaStream;
					resolve();
				},
				error => {
					reject('Error accessing local media');
				}
			);
		});
	}

	connect(userName: string) {
		this.websocket = new WebSocket('wss://' + location.host + '/match');
		this.websocket.onopen = open => {
			this.wsSend({
				name: userName,
			});
		};
		this.websocket.onclose = close => {
			this.msgSubject.complete();
		};
		this.websocket.onerror = error => {
			this.msgSubject.error(error);
		};
		this.websocket.onmessage = messageEvent => {
			let message: any;
			try {
				message = JSON.parse(messageEvent.data);
			} catch (e) {
				this.sendNext();
				this.msgSubject.next(new EndMessage());
			}
			if (message.type == 'start') {
				this.matchID = message.matchID;
				this.peerName = message.peerName;
				this.msgSubject.next(new StartMessage(this.peerName));
			} else if (message.matchID == this.matchID) {
				switch (message.type) {
				case 'end':
					this.sendNext();
					this.msgSubject.next(new EndMessage());
					break;
				}
			}
		};
	}

	private wsSend(msg: any) {
		this.websocket.send(JSON.stringify(msg));
	}

	sendNext() {
		if (this.matchID != null) {
			this.wsSend({
				type: 'next',
				matchID: this.matchID,
			});
			this.matchID = null;
		}
	}
}
