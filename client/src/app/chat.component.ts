import {Component, EventEmitter, Input, Output, OnInit, ChangeDetectorRef} from '@angular/core';
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
	localStreamURL: SafeUrl;
	remoteStreamURL: SafeUrl;
	peerName: string = null;

	constructor(
		private nextService: NextService,
		private changeDetector: ChangeDetectorRef,
		private sanitizer: DomSanitizationService
	) {}

	ngOnInit() {
		this.localStreamURL = this.sanitizer.bypassSecurityTrustUrl(this.nextService.getLocalStreamURL());
		this.setRemoteStreamURL('');

		this.nextService.connect(this.userName);

		this.nextService.msgObservable.subscribe(
			msg => {
				if (msg instanceof StartMessage) {
					this.peerName = msg.peerName;
					this.setRemoteStreamURL(msg.remoteStreamURL);
					this.changeDetector.detectChanges();
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
		this.setRemoteStreamURL('');
	}

	private setRemoteStreamURL(url: string) {
		this.remoteStreamURL = this.sanitizer.bypassSecurityTrustUrl(url);
	}
}
