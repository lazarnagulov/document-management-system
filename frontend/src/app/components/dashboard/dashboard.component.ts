import {Component, inject, signal, computed, OnInit, OnDestroy} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../core/services/auth.service';
import { DocumentService } from '../../core/services/document.service';
import { NavbarComponent } from '../../shared/navbar/navbar.component';
import { ModalShellComponent } from '../../shared/modal-shell/modal-shell.component';
import { PaginationComponent, PageEvent } from '../../shared/pagination/pagination.component';
import { DocumentCardComponent } from './document-card/document-card.component';
import { ChatMessage, DocumentResponse } from '../../models/document.model';
import { PAGE_SIZE, PageResponse } from '../../models/pagination.model';
import {PollingService} from '../../core/services/polling.service';

type ModalMode = 'upload' | 'rename' | 'delete' | 'detail' | null;

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    NavbarComponent,
    ModalShellComponent,
    PaginationComponent,
    DocumentCardComponent,
  ],
  templateUrl: './dashboard.component.html'
})
export class DashboardComponent implements OnInit, OnDestroy {
  private readonly authService     = inject(AuthService);
  private readonly documentService = inject(DocumentService);
  private readonly pollingService  = inject(PollingService);

  pagedResponse = signal<PageResponse<DocumentResponse>>({
    content: [], page: 1, size: PAGE_SIZE, totalElements: 0, totalPages: 0,
  });
  currentPage   = signal(1);
  pageSize      = PAGE_SIZE;
  isListLoading = signal(false);
  listError     = signal<string | null>(null);

  selectedDocument = signal<DocumentResponse | null>(null);
  messages         = signal<ChatMessage[]>([]);
  currentQuery     = '';
  isAiLoading      = computed(() => this.messages().some(m => m.loading));

  modalMode      = signal<ModalMode>(null);
  targetDocument = signal<DocumentResponse | null>(null);
  renameValue    = '';
  dragOver       = false;

  uploadFile:  File | null = null;
  uploadError: string      = '';
  isUploading: boolean     = false;

  ngOnInit() {
    this.loadPage(1);
  }

  ngOnDestroy() {
    this.pollingService.stop();
  }

  loadPage(page: number) {
    this.isListLoading.set(true);
    this.listError.set(null);

    this.documentService.list(page, this.pageSize).subscribe({
      next: response => {
        this.pagedResponse.set(response);
        this.currentPage.set(response.page);
        this.isListLoading.set(false);
        this.pollingService.watch(this.pagedResponse, this.selectedDocument);
      },
      error: () => {
        this.listError.set('Failed to load documents. Please try again.');
        this.isListLoading.set(false);
      },
    });
  }
  onPageChange(event: PageEvent) {
    this.loadPage(event.page);
  }

  logout() {
    this.authService.logout();
  }

  selectDocument(doc: DocumentResponse) {
    if (this.selectedDocument()?.id !== doc.id) {
      this.selectedDocument.set(doc);
      this.messages.set([]);
      this.currentQuery = '';
    }
  }

  openUploadModal() {
    this.uploadFile  = null;
    this.uploadError = '';
    this.modalMode.set('upload');
  }

  onDragOver(event: DragEvent) {
    event.preventDefault();
    this.dragOver = true;
  }

  onDragLeave() {
    this.dragOver = false;
  }

  onDrop(event: DragEvent) {
    event.preventDefault();
    this.dragOver = false;
    const file = event.dataTransfer?.files[0];
    if (file) this.setUploadFile(file);
  }

  onFileSelected(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) this.setUploadFile(file);
  }

  private setUploadFile(file: File) {
    const allowed = [
      'application/pdf',
      'text/plain',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    ];
    if (!allowed.includes(file.type)) {
      this.uploadError = 'Only PDF, DOCX, and TXT files are supported.';
      this.uploadFile  = null;
      return;
    }
    if (file.size > 50 * 1024 * 1024) {
      this.uploadError = 'File must be under 50MB.';
      this.uploadFile  = null;
      return;
    }
    this.uploadError = '';
    this.uploadFile  = file;
  }

  confirmUpload() {
    if (!this.uploadFile) return;
    this.isUploading = true;

    this.documentService.create(this.uploadFile).subscribe({
      next: () => {
        this.isUploading = false;
        this.closeModal();
        this.loadPage(1);
      },
      error: () => {
        this.isUploading = false;
        this.uploadError = 'Upload failed. Please try again.';
      },
    });
  }

  openRenameModal(doc: DocumentResponse) {
    this.targetDocument.set(doc);
    this.renameValue = doc.name;
    this.modalMode.set('rename');
  }

  confirmRename() {
    const doc = this.targetDocument();
    if (!doc || !this.renameValue.trim()) return;

    this.documentService.update(doc.id, { name: this.renameValue.trim() }).subscribe({
      next: updated => {
        this.pagedResponse.update(r => ({
          ...r,
          content: r.content.map(d => d.id === updated.id ? updated : d),
        }));
        if (this.selectedDocument()?.id === updated.id) {
          this.selectedDocument.set(updated);
        }
        this.closeModal();
      },
      error: () => {
        this.closeModal();
        this.loadPage(this.currentPage());
      },
    });
  }

  openDeleteModal(doc: DocumentResponse) {
    this.targetDocument.set(doc);
    this.modalMode.set('delete');
  }

  openDetailModal(doc: DocumentResponse) {
    this.targetDocument.set(doc);
    this.modalMode.set('detail');
  }

  confirmDelete() {
    const doc = this.targetDocument();
    if (!doc) return;

    this.documentService.delete(doc.id).subscribe({
      next: () => {
        if (this.selectedDocument()?.id === doc.id) {
          this.selectedDocument.set(null);
          this.messages.set([]);
        }
        this.closeModal();

        const isLastOnPage = this.pagedResponse().content.length === 1 && this.currentPage() > 1;
        this.loadPage(isLastOnPage ? this.currentPage() - 1 : this.currentPage());
      },
      error: () => {
        this.closeModal();
        this.loadPage(this.currentPage());
      },
    });
  }

  downloadDocument(doc: DocumentResponse) {
    this.documentService.download(doc.id, doc.name).subscribe();
  }

  closeModal() {
    this.modalMode.set(null);
    this.targetDocument.set(null);
    this.uploadFile  = null;
    this.uploadError = '';
    this.dragOver    = false;
  }

  askQuestion() {
    const query = this.currentQuery.trim();
    const doc   = this.selectedDocument();
    if (!query || !doc || this.isAiLoading()) return;

    this.messages.update(msgs => [
      ...msgs,
      { sender: 'user', text: query },
      { sender: 'ai',   text: '', loading: true },
    ]);
    this.currentQuery = '';

    this.documentService.query(doc.id, query).subscribe({
      next: response => {
        this.messages.update(msgs => [
          ...msgs.filter(m => !m.loading),
          { sender: 'ai', text: response.answer },
        ]);
      },
      error: () => {
        this.messages.update(msgs => [
          ...msgs.filter(m => !m.loading),
          { sender: 'ai', text: 'Something went wrong. Please try again.' },
        ]);
      },
    });
  }

  formatSize(bytes: number): string {
    if (bytes < 1024)        return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
}
