import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { DocumentList } from '../document-list/document-list';

@Component({
  selector: 'app-dashboard',
  imports: [CommonModule, RouterModule, DocumentList],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard implements OnInit {
  title = 'lucidRAG Admin Dashboard';

  ngOnInit(): void {
    console.log('Dashboard initialized');
  }
}
