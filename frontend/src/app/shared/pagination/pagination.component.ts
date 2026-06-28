import { Component, EventEmitter, Input, OnChanges, Output } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface PageEvent {
  page: number;
  size: number;
}

@Component({
  selector: 'app-pagination',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './pagination.component.html'
})
export class PaginationComponent implements OnChanges {
  @Input() page = 1;
  @Input() totalPages = 0;
  @Input() totalElements = 0;
  @Input() size = 10;

  @Output() pageChange = new EventEmitter<PageEvent>();

  pages: number[] = [];

  ngOnChanges() {
    this.buildPages();
  }

  private buildPages() {
    if (this.totalPages <= 7) {
      this.pages = Array.from({ length: this.totalPages }, (_, i) => i + 1);
      return;
    }
    const around = new Set([
      1,
      this.totalPages,
      this.page,
      this.page - 1,
      this.page + 1,
    ].filter(p => p >= 1 && p <= this.totalPages));

    this.pages = Array.from(around).sort((a, b) => a - b);
  }

  isEllipsisBefore(index: number): boolean {
    return index > 0 && this.pages[index] - this.pages[index - 1] > 1;
  }

  go(page: number) {
    if (page < 1 || page > this.totalPages || page === this.page) return;
    this.pageChange.emit({ page, size: this.size });
  }

  get hasPrev(): boolean { return this.page > 1; }
  get hasNext(): boolean { return this.page < this.totalPages; }
}
