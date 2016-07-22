import {Component, EventEmitter, Input, Output, OnInit} from '@angular/core';
import {DomSanitizationService, SafeUrl} from '@angular/platform-browser';
import {MdButton} from '@angular2-material/button';
import {MdProgressBar} from '@angular2-material/progress-bar';
import {NextService, StartMessage, EndMessage} from './next.service';

@Component({
	selector: 'chat',
	template: require('./chat.component.html'),
	styles: [require('./chat.component.css')],
	directives: [MdButton, MdProgressBar],
})
export class ChatComponent implements OnInit {
	@Input() userName: string;
	@Output() onError = new EventEmitter<string>();
	peerName: string = null;

	localStreamURL: SafeUrl;

	constructor(
		private nextService: NextService,
		private sanitizer: DomSanitizationService
	) {}

	ngOnInit() {
		this.localStreamURL = this.sanitizer.bypassSecurityTrustUrl(this.nextService.getLocalStreamURL());

		this.nextService.connect(this.userName);

		this.nextService.msgObservable.subscribe(
			msg => {
				if (msg instanceof StartMessage) {
					this.peerName = msg.peerName;
				} else if (msg instanceof EndMessage) {
					this.resetMatch();
				}
			},
			error => {
				this.onError.emit('An error occured with WebSocket connection');
			},
			() => {
				this.onError.emit('WebSocket has been closed');
			}
		);
	}

	next() {
		this.resetMatch();
		this.nextService.sendNext();
	}

	resetMatch() {
		this.peerName = null;
	}
}
