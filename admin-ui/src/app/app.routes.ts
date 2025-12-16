import { Routes } from '@angular/router';
import { Dashboard } from './components/dashboard/dashboard';
import { DocumentForm } from './components/document-form/document-form';
import { ConversationList } from './components/conversation-list/conversation-list';
import { ConversationDetail } from './components/conversation-detail/conversation-detail';

export const routes: Routes = [
  { path: '', component: Dashboard },
  { path: 'documents/new', component: DocumentForm },
  { path: 'conversations', component: ConversationList },
  { path: 'conversations/:id', component: ConversationDetail },
  { path: '**', redirectTo: '' }
];
