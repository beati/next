import {Component, EventEmitter, Output} from '@angular/core';

@Component({
	selector: 'register',
	template: require('./register.component.html'),
	styles: [`
		.container {
		box-sizing: border-box;
		width: 100%;
		max-width: 500px;
		margin: 0 auto;
		padding: 5px;
		margin-top: 50px;
		}
	`],
})
export class RegisterComponent {
	@Output() onRegister = new EventEmitter<string>();

	register(userName: string) {
		this.onRegister.emit(userName);
	}
}
