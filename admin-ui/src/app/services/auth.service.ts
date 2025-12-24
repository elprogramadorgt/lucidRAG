import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap, catchError, throwError } from 'rxjs';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';
import {
  User,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  AuthState
} from '../models/user.model';

const TOKEN_KEY = 'lucidrag_token';
const USER_KEY = 'lucidrag_user';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = `${environment.apiUrl}/auth`;

  private authState = signal<AuthState>({
    user: null,
    token: null,
    isAuthenticated: false
  });

  readonly user = computed(() => this.authState().user);
  readonly token = computed(() => this.authState().token);
  readonly isAuthenticated = computed(() => this.authState().isAuthenticated);
  readonly isAdmin = computed(() => this.authState().user?.role === 'admin');

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    this.loadFromStorage();
  }

  private loadFromStorage(): void {
    const token = localStorage.getItem(TOKEN_KEY);
    const userJson = localStorage.getItem(USER_KEY);

    if (token && userJson) {
      try {
        const user = JSON.parse(userJson) as User;
        this.authState.set({
          user,
          token,
          isAuthenticated: true
        });
      } catch {
        this.clearStorage();
      }
    }
  }

  private saveToStorage(token: string, user: User): void {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }

  private clearStorage(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }

  login(credentials: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.apiUrl}/login`, credentials).pipe(
      tap(response => {
        this.saveToStorage(response.token, response.user);
        this.authState.set({
          user: response.user,
          token: response.token,
          isAuthenticated: true
        });
      }),
      catchError(error => {
        return throwError(() => error);
      })
    );
  }

  register(data: RegisterRequest): Observable<RegisterResponse> {
    return this.http.post<RegisterResponse>(`${this.apiUrl}/register`, data);
  }

  logout(): void {
    this.clearStorage();
    this.authState.set({
      user: null,
      token: null,
      isAuthenticated: false
    });
    this.router.navigate(['/login']);
  }

  getToken(): string | null {
    return this.authState().token;
  }

  getCurrentUser(): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/me`).pipe(
      tap(user => {
        const token = this.authState().token;
        if (token) {
          this.saveToStorage(token, user);
          this.authState.set({
            user,
            token,
            isAuthenticated: true
          });
        }
      })
    );
  }
}
