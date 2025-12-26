import { Routes } from '@angular/router';
import { LoginComponent } from './components/login/login';
import { RegisterComponent } from './components/register/register';
import { OAuthCallbackComponent } from './components/oauth-callback/oauth-callback';
import { DashboardComponent } from './components/dashboard/dashboard';
import { ConversationsPageComponent } from './components/conversations-page/conversations-page';
import { DocumentsPageComponent } from './components/documents-page/documents-page';
import { SystemLogsComponent } from './components/system-logs/system-logs';
import { authGuard } from './guards/auth.guard';
import { adminGuard } from './guards/admin.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'dashboard',
    pathMatch: 'full',
  },
  {
    path: 'login',
    component: LoginComponent,
  },
  {
    path: 'register',
    component: RegisterComponent,
  },
  {
    path: 'oauth/callback',
    component: OAuthCallbackComponent,
  },
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [authGuard],
  },
  {
    path: 'conversations',
    component: ConversationsPageComponent,
    canActivate: [authGuard],
  },
  {
    path: 'documents',
    component: DocumentsPageComponent,
    canActivate: [authGuard],
  },
  {
    path: 'system',
    component: SystemLogsComponent,
    canActivate: [adminGuard],
  },
  {
    path: '**',
    redirectTo: 'dashboard',
  },
];
