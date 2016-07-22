import {Component, OnInit} from '@angular/core';
import {MdButton} from '@angular2-material/button';
import {MdCard} from '@angular2-material/card';
import {NextService} from './next.service';
import {RegisterComponent} from './register.component';
import {MediaAccessComponent} from './mediaAccess.component';
import {ChatComponent} from './chat.component';

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
	directives: [MdButton, MdCard, RegisterComponent, MediaAccessComponent, ChatComponent],
	providers: [NextService],
})
export class AppComponent implements OnInit {
	initLevel: string;
	errorMessage = '';
	userName: string;

	ngOnInit() {
		this.initLevel = 'notsupported';
		switch (adapter.browserDetails.browser) {
		case 'chrome':
		case 'firefox':
			if (adapter.browserDetails.version > adapter.browserDetails.minVersion) {
				this.initLevel = 'registering';
			}
			break;
		}
	}

	onRegister(userName: string) {
		this.userName = userName;
		this.initLevel = 'accessing_media';
	}

	onAccessed() {
		this.initLevel = 'chatting';
	}

	onError(errorMessage: string) {
		this.errorMessage = errorMessage;
		this.initLevel = 'error';
	}

	reload() {
		location.reload(true);
	}
}
