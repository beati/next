import {Component, EventEmitter, Output, OnInit} from '@angular/core';
import {MdCard} from '@angular2-material/card';
import {NextService} from './next.service';

@Component({
	selector: 'mediaAccess',
	template:`
		<div class="container">
			<md-card>Accessing local media</md-card>
		<div>
	`,
	styles: [`
		.container {
		box-sizing: border-box;
		width: 100%;
		max-width: 500px;
		margin: 0 auto;
		padding: 5px;
		margin-top: 200px;
		}
	`],
	directives: [MdCard],
})
export class MediaAccessComponent implements OnInit {
	@Output() onAccessed = new EventEmitter<void>();
	@Output() onError = new EventEmitter<string>();

	constructor(private nextService: NextService) {
	}

	ngOnInit() {
		this.nextService.getUserMedia().then(() => {
			this.onAccessed.emit(null);
		}).catch(error => {
			this.onError.emit(error);
		});
	}
}
