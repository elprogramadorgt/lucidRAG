import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { DocumentList } from '../document-list/document-list';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-dashboard',
  imports: [CommonModule, RouterModule, DocumentList],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard implements OnInit {
  title = 'lucidRAG Admin Dashboard';

  constructor(public authService: AuthService) {}

  ngOnInit(): void {
    console.log('Dashboard initialized');
  }

  logout(): void {
    this.authService.logout();
  }
}
