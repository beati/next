import {Injectable} from '@angular/core';
import {Subject} from 'rxjs/Subject';

export class StartMessage {
	constructor(
		public peerName: string,
		public remoteStreamUrl: string
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
				this.sendNext();
				this.msgSubject.next(new EndMessage());
			}
			if (message.type == 'start') {
				this.matchID = message.matchID;
				this.peerName = message.peerName;
				this.offer = message.offer;
				this.startWebrtcConnection();
			} else if (message.matchID == this.matchID) {
				switch (message.type) {
				case 'end':
					this.sendNext();
					this.msgSubject.next(new EndMessage());
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

	startWebrtcConnection() {
		let config: any = {
			iceServers: [
				{
					urls: 'stun:stun.l.google.com:19302',
				},
			],
		};
		this.peerConnection = new RTCPeerConnection(config);

		this.peerConnection.addStream(this.localStream);

		this.peerConnection.onaddstream = evt => {
			let remoteStreamUrl = URL.createObjectURL(evt.stream);
			this.msgSubject.next(new StartMessage(
				this.peerName,
				remoteStreamUrl
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
		}

		if (this.offer) {
			this.peerConnection.createOffer(
				sdp => {
					this.setLocalDescription(sdp);
				},
				error => {
					console.log(error);
				}
			);
		}
	}

	private receiveSDP(sdp: RTCSessionDescription) {
		let pc = this.peerConnection;
		pc.setRemoteDescription(
			sdp,
			() => {
				if (!this.offer) {
					pc.createAnswer(
						sdp => {
							this.setLocalDescription(sdp);
						},
						error => {
							console.log(error);
						}
					);
				}
			},
			error => {
				console.log(error);
			}
		);
	}

	private setLocalDescription(sdp: RTCSessionDescription) {
		this.peerConnection.setLocalDescription(
			sdp,
			() => {
				this.wsSend({
					type: 'sdp',
					matchID: this.matchID,
					data: sdp,
				});
			},
			error => {
				console.log(error);
			}
		);
	}

	private receiveCandidate(candidate: RTCIceCandidate) {
		this.peerConnection.addIceCandidate(
			candidate,
			() => {},
			error => {
				console.log(error);
			}
		);
	}
}
