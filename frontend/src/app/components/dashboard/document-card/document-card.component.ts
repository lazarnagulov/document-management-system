import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DocumentResponse } from '../../../models/document.model';

@Component({
  selector: 'app-document-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './document-card.component.html'
})
export class DocumentCardComponent {
  @Input({ required: true }) document!: DocumentResponse;
  @Input() isSelected = false;

  @Output() selectDoc    = new EventEmitter<DocumentResponse>();
  @Output() renameDoc    = new EventEmitter<DocumentResponse>();
  @Output() deleteDoc    = new EventEmitter<DocumentResponse>();
  @Output() downloadDoc  = new EventEmitter<DocumentResponse>();

  get canSelect(): boolean {
    return this.document.status === 'READY';
  }

  onClick() {
    if (this.canSelect) this.selectDoc.emit(this.document);
  }

  onRename(e: Event)   { e.stopPropagation(); this.renameDoc.emit(this.document); }
  onDelete(e: Event)   { e.stopPropagation(); this.deleteDoc.emit(this.document); }
  onDownload(e: Event) { e.stopPropagation(); this.downloadDoc.emit(this.document); }
}
