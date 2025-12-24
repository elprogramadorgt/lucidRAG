import { Routes } from '@angular/router';
import { Dashboard } from './components/dashboard/dashboard';
import { DocumentForm } from './components/document-form/document-form';
import { Login } from './components/login/login';
import { ConversationList } from './components/conversation-list/conversation-list';
import { ConversationDetail } from './components/conversation-detail/conversation-detail';
import { authGuard, adminGuard, guestGuard } from './guards/auth.guard';

export const routes: Routes = [
  { path: 'login', component: Login, canActivate: [guestGuard] },
  { path: '', component: Dashboard, canActivate: [authGuard] },
  { path: 'documents/new', component: DocumentForm, canActivate: [adminGuard] },
  { path: 'conversations', component: ConversationList, canActivate: [adminGuard] },
  { path: 'conversations/:id', component: ConversationDetail, canActivate: [adminGuard] },
  { path: '**', redirectTo: '' }
];
