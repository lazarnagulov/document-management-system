import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import {PageResponse} from '../../models/pagination.model';
import {DocumentResponse, QueryRequest, QueryResponse, UpdateDocumentRequest} from '../../models/document.model';

@Injectable({ providedIn: 'root' })
export class DocumentService {
  private readonly http = inject(HttpClient);
  private readonly API  = '/api/v1/documents';

  list(page: number, size: number): Observable<PageResponse<DocumentResponse>> {
    const params = new HttpParams()
      .set('page', page)
      .set('size', size);
    return this.http.get<PageResponse<DocumentResponse>>(`${this.API}/`, { params });
  }

  get(id: string): Observable<DocumentResponse> {
    return this.http.get<DocumentResponse>(`${this.API}/${id}`);
  }

  create(file: File): Observable<DocumentResponse> {
    const form = new FormData();
    form.append('document', file);
    return this.http.post<DocumentResponse>(`${this.API}/`, form);
  }

  update(id: string, req: UpdateDocumentRequest): Observable<DocumentResponse> {
    return this.http.patch<DocumentResponse>(`${this.API}/${id}`, req);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.API}/${id}`);
  }

  download(id: string, filename: string): Observable<void> {
    return new Observable(observer => {
      this.http.get(`${this.API}/${id}/download`, { responseType: 'blob' }).subscribe({
        next: blob => {
          const url = URL.createObjectURL(blob);
          const a   = document.createElement('a');
          a.href     = url;
          a.download = filename;
          a.click();
          URL.revokeObjectURL(url);
          observer.next();
          observer.complete();
        },
        error: err => observer.error(err),
      });
    });
  }

  query(id: string, question: string): Observable<QueryResponse> {
    return this.http.post<QueryResponse>(`${this.API}/${id}/query`, { question } as QueryRequest);
  }
}
