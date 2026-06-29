import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-modal-shell',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './modal-shell.component.html'
})
export class ModalShellComponent {
  @Input() title = '';
  @Output() closeClick = new EventEmitter<void>();
}
