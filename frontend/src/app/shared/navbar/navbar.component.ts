import { Component, EventEmitter, Input, Output } from '@angular/core';
import {CommonModule, NgOptimizedImage} from '@angular/common';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule, NgOptimizedImage],
  templateUrl: './navbar.component.html'
})
export class NavbarComponent {
  @Input() title = 'DMS Intelligence';
  @Output() logoutClick = new EventEmitter<void>();
}
