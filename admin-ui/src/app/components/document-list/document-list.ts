import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DocumentService } from '../../services/document.service';
import { Document } from '../../models/document.model';

@Component({
  selector: 'app-document-list',
  imports: [CommonModule],
  templateUrl: './document-list.html',
  styleUrl: './document-list.scss',
})
export class DocumentList implements OnInit {
  documents: Document[] = [];
  loading = false;
  error: string | null = null;

  constructor(private documentService: DocumentService) {}

  ngOnInit(): void {
    this.loadDocuments();
  }

  loadDocuments(): void {
    this.loading = true;
    this.error = null;
    
    this.documentService.getDocuments(10, 0).subscribe({
      next: (response) => {
        this.documents = response.documents || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load documents: ' + err.message;
        this.loading = false;
      }
    });
  }

  deleteDocument(id: string): void {
    if (!confirm('Are you sure you want to delete this document?')) {
      return;
    }

    this.documentService.deleteDocument(id).subscribe({
      next: () => {
        this.loadDocuments();
      },
      error: (err) => {
        this.error = 'Failed to delete document: ' + err.message;
      }
    });
  }
}
