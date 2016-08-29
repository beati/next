import {Component, OnInit, ChangeDetectorRef} from '@angular/core';

declare var adapter: any;

@Component({
	selector: 'main',
	template: require('./app.component.html'),
	styles: [`
		.container {
		box-sizing: border-box;
		width: 100%;
		max-width: 500px;
		margin: 0 auto;
		padding: 5px;
		margin-top: 200px;
		}

		.error-message {
		padding-right: 70px;
		}
	`],
})
export class AppComponent implements OnInit {
	initLevel: string;
	errorMessage = '';
	userName: string;

	constructor(private changeDetector: ChangeDetectorRef) {}

	ngOnInit() {
		this.initLevel = 'notsupported';
		switch (adapter.browserDetails.browser) {
		case 'chrome':
		case 'firefox':
			this.initLevel = 'registering';
			break;
		}
	}

	onRegister(userName: string) {
		this.userName = userName;
		this.initLevel = 'accessing_media';
	}

	onAccessed() {
		this.initLevel = 'chatting';
		this.changeDetector.detectChanges();
	}

	onError(errorMessage: string) {
		this.errorMessage = errorMessage;
		this.initLevel = 'error';
		this.changeDetector.detectChanges();
	}

	reload() {
		location.reload(true);
	}
}
