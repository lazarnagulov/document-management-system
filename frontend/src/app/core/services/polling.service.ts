import {inject, Injectable, OnDestroy, WritableSignal} from '@angular/core';
import {DocumentResponse} from '../../models/document.model';
import {DocumentService} from './document.service';
import {PageResponse} from '../../models/pagination.model';

const POLL_MS = 3000;
const PENDING: DocumentResponse['status'][] = ['UPLOADED', 'PROCESSING'];

@Injectable({
  providedIn: 'root'
})
export class PollingService implements OnDestroy {
  private readonly documentService = inject(DocumentService);
  private interval: ReturnType<typeof setInterval> | null = null;

  watch(
    pagedResponse: WritableSignal<PageResponse<DocumentResponse>>,
    selectedDocument: WritableSignal<DocumentResponse | null>,
  ) {
    this.stop();
    this.checkAndSchedule(pagedResponse, selectedDocument);
  }

  private checkAndSchedule(
    pagedResponse: WritableSignal<PageResponse<DocumentResponse>>,
    selectedDocument: WritableSignal<DocumentResponse | null>,
  ) {
    const hasPending = pagedResponse().content.some(d => PENDING.includes(d.status));

    if (!hasPending || this.interval) return;

    this.interval = setInterval(() => {
      this.tick(pagedResponse, selectedDocument);
    }, POLL_MS);
  }

  private tick(
    pagedResponse: WritableSignal<PageResponse<DocumentResponse>>,
    selectedDocument: WritableSignal<DocumentResponse | null>,
  ) {
    const pending = pagedResponse().content.filter(d => PENDING.includes(d.status));

    if (pending.length === 0) {
      this.stop();
      return;
    }

    pending.forEach(doc => {
      this.documentService.get(doc.id).subscribe({
        next: updated => {
          if (updated.status === doc.status) return;

          pagedResponse.update(r => ({
            ...r,
            content: r.content.map(d => d.id === updated.id ? updated : d),
          }));

          if (selectedDocument()?.id === updated.id) {
            selectedDocument.set(updated);
          }

          const stillPending = pagedResponse().content.some(d => PENDING.includes(d.status));
          if (!stillPending) this.stop();
        },
      });
    });
  }

  stop() {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
  }

  ngOnDestroy() {
    this.stop();
  }


  constructor() { }
}
