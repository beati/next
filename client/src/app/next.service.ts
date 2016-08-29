import {Injectable} from '@angular/core';
import {Subject} from 'rxjs/Subject';

export class StartMessage {
	constructor(
		public peerName: string,
		public remoteStreamURL: string
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

	private peerConnection: RTCPeerConnection;
	private offer: boolean;

	getUserMedia(): Promise<void> {
		return navigator.mediaDevices.getUserMedia({
			audio: true,
			video: true,
		})
		.then(stream => {
			this.localStream = stream;
		})
		.catch(error => {
			let errorMessage = '';
			console.log(error.name);
			switch (error.name) {
			case 'NotAllowedError':
				errorMessage = 'Access to media refused';
				break;
			case 'DevicesNotFoundError':
			case 'NotFoundError':
				errorMessage = 'No media device found';
				break;
			case 'SourceUnavailableError':
				errorMessage = 'Media devices allready in use';
				break;
			}
			throw errorMessage;
		});
	}

	getLocalStreamURL(): string {
		let url = '';
		if (this.localStream != null) {
			url = URL.createObjectURL(this.localStream);
		}
		return url;
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
				this.endMatch();
			}
			if (message.type == 'start') {
				this.matchID = message.matchID;
				this.peerName = message.peerName;
				this.offer = message.offer;
				if (!(message.turnUsername && message.turnPassword)) {
					message.turnUsername = "";
					message.turnPassword = "";
				}
				this.startWebrtcConnection(message.turnUsername, message.turnPassword);
			} else if (message.matchID == this.matchID) {
				switch (message.type) {
				case 'end':
					this.endMatch();
					break;
				case 'sdp':
					this.receiveSDP(message.data);
					break;
				case 'candidate':
					this.receiveCandidate(message.data);
					break;
				}
			}
		};
	}

	private wsSend(msg: any) {
		this.websocket.send(JSON.stringify(msg));
	}

	sendNext() {
		this.peerConnection.close();
		if (this.matchID != null) {
			this.wsSend({
				type: 'next',
				matchID: this.matchID,
			});
			this.matchID = null;
		}
	}

	endMatch() {
		this.sendNext();
		this.msgSubject.next(new EndMessage());
	}

	private startWebrtcConnection(turnUsername: string, turnPassword: string) {
		let config: RTCConfiguration = {
			iceServers: [
				{
					urls: 'stun:stun.l.google.com:19302',
				},
			],
		};
		if (turnUsername != "" && turnPassword != "") {
			config.iceServers.push({
				urls: 'turn:next.beati.io:3478?transport=udp',
				username: turnUsername,
				credential: turnPassword,
			});
		}
		this.peerConnection = new RTCPeerConnection(config);

		this.peerConnection.oniceconnectionstatechange = evt => {
			if (this.peerConnection.iceConnectionState == 'failed') {
				this.endMatch();
			}
		};

		this.peerConnection.onaddstream = evt => {
			let remoteStreamURL = URL.createObjectURL(evt.stream);
			this.msgSubject.next(new StartMessage(
				this.peerName,
				remoteStreamURL
			));
		};

		this.peerConnection.onicecandidate = evt => {
			if (this.matchID != null) {
				if (evt.candidate) {
					this.wsSend({
						type: 'candidate',
						matchID: this.matchID,
						data: evt.candidate,
					});
				}
			}
		};

		this.peerConnection.addStream(this.localStream);

		if (this.offer) {
			let sdpOffer = this.peerConnection.createOffer()
			this.sendSDP(sdpOffer);
		}
	}

	private sendSDP(sdp: Promise<RTCSessionDescription>) {
		sdp
		.then(sdp => {
			return this.peerConnection.setLocalDescription(sdp);
		})
		.then(() => {
			this.wsSend({
				type: 'sdp',
				matchID: this.matchID,
				data: this.peerConnection.localDescription,
			});
		})
		.catch(error => {
			this.endMatch();
		});
	}

	private receiveSDP(sdp: RTCSessionDescription) {
		this.peerConnection.setRemoteDescription(sdp)
		.then(() => {
			if (!this.offer) {
				let sdpAnswer = this.peerConnection.createAnswer();
				this.sendSDP(sdpAnswer);
			}
		})
		.catch(error => {
			this.endMatch();
		});
	}

	private receiveCandidate(candidate: RTCIceCandidate) {
		this.peerConnection.addIceCandidate(candidate)
		.catch(error => {
			this.endMatch();
		});
	}
}
