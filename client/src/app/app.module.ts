import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {FormsModule} from '@angular/forms';

import {MdButtonModule} from '@angular2-material/button';
import {MdCardModule} from '@angular2-material/card';
import {MdInputModule} from '@angular2-material/input';
import {MdProgressBarModule} from '@angular2-material/progress-bar';

import {NextService} from './next.service';

import {AppComponent} from './app.component';
import {RegisterComponent} from './register.component';
import {MediaAccessComponent} from './mediaAccess.component';
import {ChatComponent} from './chat.component';

@NgModule({
	imports: [
		BrowserModule,
		FormsModule,

		MdButtonModule,
		MdCardModule,
		MdInputModule,
		MdProgressBarModule,
	],
	declarations: [
		AppComponent,
		RegisterComponent,
		MediaAccessComponent,
		ChatComponent,
	],
	bootstrap: [AppComponent],
	providers: [NextService],
})
export class AppModule {}
