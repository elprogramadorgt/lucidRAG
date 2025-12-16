import { Routes } from '@angular/router';
import { Dashboard } from './components/dashboard/dashboard';
import { DocumentForm } from './components/document-form/document-form';

export const routes: Routes = [
  { path: '', component: Dashboard },
  { path: 'documents/new', component: DocumentForm },
  { path: '**', redirectTo: '' }
];
