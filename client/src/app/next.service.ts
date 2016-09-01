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
				this.endMatch();
			}
			if (message.type == 'start') {
				this.matchID = message.matchID;
				this.peerName = message.peerName;
				this.msgSubject.next(new StartMessage(this.peerName));
			} else if (message.matchID == this.matchID) {
				switch (message.type) {
				case 'end':
					this.endMatch();
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

	private endMatch() {
		this.sendNext();
		this.msgSubject.next(new EndMessage());
	}
}
