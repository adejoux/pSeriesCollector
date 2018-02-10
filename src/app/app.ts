import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AppRoutes } from './app.routes';

@Component({
  selector: 'pseriescollector-app',
  templateUrl: './app.html'
})

export class App {
  constructor(public router: Router) {}
}
