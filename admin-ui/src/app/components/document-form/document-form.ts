import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DocumentService } from '../../services/document.service';
import { Document } from '../../models/document.model';

@Component({
  selector: 'app-document-form',
  imports: [CommonModule, FormsModule],
  templateUrl: './document-form.html',
  styleUrl: './document-form.scss',
})
export class DocumentForm {
  document: Partial<Document> = {
    title: '',
    content: '',
    source: '',
    is_active: true
  };

  submitting = false;
  success: string | null = null;
  error: string | null = null;

  constructor(private documentService: DocumentService) {}

  onSubmit(): void {
    this.submitting = true;
    this.success = null;
    this.error = null;

    this.documentService.createDocument(this.document).subscribe({
      next: (response) => {
        this.success = response.message;
        this.submitting = false;
        this.resetForm();
      },
      error: (err) => {
        this.error = 'Failed to create document: ' + err.message;
        this.submitting = false;
      }
    });
  }

  resetForm(): void {
    this.document = {
      title: '',
      content: '',
      source: '',
      is_active: true
    };
  }
}
