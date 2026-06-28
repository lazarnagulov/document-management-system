import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../core/services/auth.service';
import { NavbarComponent } from '../../shared/navbar/navbar.component';
import { ModalShellComponent } from '../../shared/modal-shell/modal-shell.component';
import { PaginationComponent, PageEvent } from '../../shared/pagination/pagination.component';
import { DocumentCardComponent } from './document-card/document-card.component';
import {ChatMessage, DocumentResponse} from '../../models/document.model';
import {PAGE_SIZE, PageResponse} from '../../models/pagination.model';

type ModalMode = 'upload' | 'rename' | 'delete' | null;

const ALL_MOCK_DOCUMENTS: DocumentResponse[] = Array.from({ length: 23 }, (_, i) => {
  const statuses: DocumentResponse['status'][] = ['READY', 'PROCESSING', 'FAILED', 'UPLOADED'];
  const names = [
    'Q3_Financial_Report.pdf',
    'Employee_Handbook_2024.docx',
    'Product_Roadmap_v2.pdf',
    'Meeting_Notes_June.txt',
    'Legal_Contract_Draft.docx',
    'Research_Paper_AI.pdf',
    'Budget_2025.txt',
    'User_Research_Findings.docx',
    'Onboarding_Guide.pdf',
    'Technical_Spec.txt',
  ];
  return {
    id: crypto.randomUUID(),
    name: names[i % names.length].replace('.', `_${i + 1}.`),
    content_type: 'application/pdf',
    size: Math.floor(Math.random() * 5_000_000) + 10_000,
    status: statuses[i % statuses.length],
    created_at: new Date(Date.now() - i * 86_400_000).toISOString(),
    updated_at: new Date(Date.now() - i * 43_200_000).toISOString(),
  };
});

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
export class DashboardComponent {
  private readonly authService = inject(AuthService);

  private readonly allDocuments = signal<DocumentResponse[]>([...ALL_MOCK_DOCUMENTS]);

  currentPage = signal(1);
  pageSize = PAGE_SIZE;

  pagedResponse = computed<PageResponse<DocumentResponse>>(() => {
    const all  = this.allDocuments();
    const page = this.currentPage();
    const size = this.pageSize;
    const start = (page - 1) * size;
    const content = all.slice(start, start + size);

    return {
      content,
      page,
      size,
      totalElements: all.length,
      totalPages: Math.ceil(all.length / size),
    };
  });

  onPageChange(event: PageEvent) {
    // Replace with: this.documentService.list(event.page, event.size).subscribe(...)
    this.currentPage.set(event.page);
  }

  selectedDocument = signal<DocumentResponse | null>(null);
  messages = signal<ChatMessage[]>([]);
  currentQuery = '';

  isAiLoading = computed(() => this.messages().some(m => m.loading));

  modalMode = signal<ModalMode>(null);
  targetDocument = signal<DocumentResponse | null>(null);
  renameValue = '';
  dragOver = false;

  uploadFile: File | null = null;
  uploadError = '';
  isUploading = false;

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
    this.uploadFile = null;
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
      this.uploadFile = null;
      return;
    }
    if (file.size > 50 * 1024 * 1024) {
      this.uploadError = 'File must be under 50MB.';
      this.uploadFile = null;
      return;
    }
    this.uploadError = '';
    this.uploadFile = file;
  }

  confirmUpload() {
    if (!this.uploadFile) return;
    this.isUploading = true;

    // Mock upload — replace with: this.documentService.create(formData).subscribe(...)
    setTimeout(() => {
      const newDoc: DocumentResponse = {
        id: crypto.randomUUID(),
        name: this.uploadFile!.name,
        content_type: this.uploadFile!.type,
        size: this.uploadFile!.size,
        status: 'UPLOADED',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      this.allDocuments.update(docs => [newDoc, ...docs]);
      this.currentPage.set(1); // jump to first page to see the new doc
      this.isUploading = false;
      this.closeModal();

      // Mock processing progression: UPLOADED → PROCESSING → READY
      setTimeout(() => {
        this.updateDoc(newDoc.id, { status: 'PROCESSING' });
        setTimeout(() => {
          this.updateDoc(newDoc.id, { status: 'READY' });
        }, 4000);
      }, 1000);
    }, 1200);
  }

  openRenameModal(doc: DocumentResponse) {
    this.targetDocument.set(doc);
    this.renameValue = doc.name;
    this.modalMode.set('rename');
  }

  confirmRename() {
    const doc = this.targetDocument();
    if (!doc || !this.renameValue.trim()) return;

    // Mock rename - replace with: this.documentService.update(doc.id, { name }).subscribe(...)
    const name = this.renameValue.trim();
    this.updateDoc(doc.id, { name });
    if (this.selectedDocument()?.id === doc.id) {
      this.selectedDocument.update(d => d ? { ...d, name } : d);
    }
    this.closeModal();
  }

  openDeleteModal(doc: DocumentResponse) {
    this.targetDocument.set(doc);
    this.modalMode.set('delete');
  }

  confirmDelete() {
    const doc = this.targetDocument();
    if (!doc) return;

    // Mock delete - replace with: this.documentService.delete(doc.id).subscribe(...)
    this.allDocuments.update(docs => docs.filter(d => d.id !== doc.id));

    const remaining = this.allDocuments().length;
    const maxPage = Math.max(1, Math.ceil(remaining / this.pageSize));
    if (this.currentPage() > maxPage) this.currentPage.set(maxPage);

    if (this.selectedDocument()?.id === doc.id) {
      this.selectedDocument.set(null);
      this.messages.set([]);
    }
    this.closeModal();
  }

  downloadDocument(doc: DocumentResponse) {
    // Mock download - replace with: this.documentService.download(doc.id)
    console.log('Downloading:', doc.name);
    alert(`Mock download: ${doc.name}`);
  }

  closeModal() {
    this.modalMode.set(null);
    this.targetDocument.set(null);
    this.uploadFile = null;
    this.uploadError = '';
    this.dragOver = false;
  }

  askQuestion() {
    const query = this.currentQuery.trim();
    const doc = this.selectedDocument();
    if (!query || !doc || this.isAiLoading()) return;

    this.messages.update(msgs => [
      ...msgs,
      { sender: 'user', text: query },
      { sender: 'ai', text: '', loading: true },
    ]);
    this.currentQuery = '';

    // Mock query - replace with: this.documentService.query(doc.id, query).subscribe(...)
    setTimeout(() => {
      this.messages.update(msgs => [
        ...msgs.filter(m => !m.loading),
        {
          sender: 'ai',
          text: `Based on "${doc.name}", here is a mock response to: "${query}". In the real implementation this calls POST /documents/${doc.id}/query.`,
        },
      ]);
    }, 1000);
  }

  private updateDoc(id: string, patch: Partial<DocumentResponse>) {
    this.allDocuments.update(docs =>
      docs.map(d => d.id === id ? { ...d, ...patch, updated_at: new Date().toISOString() } : d)
    );
  }

  formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
}
