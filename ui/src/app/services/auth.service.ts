import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap, catchError, of, map } from 'rxjs';
import { LoginRequest, RegisterRequest, LoginResponse, User } from '../models/user.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly USER_KEY = 'lucidrag_user';

  currentUser = signal<User | null>(null);
  isAuthenticated = signal<boolean>(false);
  isInitialized = signal<boolean>(false);

  constructor(private http: HttpClient, private router: Router) {
    this.initSession();
  }

  login(credentials: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${environment.apiUrl}/v1/auth/login`, credentials).pipe(
      tap((response) => {
        this.setUser(response.user);
      })
    );
  }

  register(data: RegisterRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${environment.apiUrl}/v1/auth/register`, data).pipe(
      tap((response) => {
        this.setUser(response.user);
      })
    );
  }

  logout(): void {
    // Call backend to clear cookie
    this.http.post(`${environment.apiUrl}/v1/auth/logout`, {}).subscribe({
      complete: () => {
        this.clearSession();
        this.router.navigate(['/login']);
      },
      error: () => {
        // Clear local state even if backend call fails
        this.clearSession();
        this.router.navigate(['/login']);
      }
    });
  }

  hasPermission(requiredRole: 'admin' | 'user'): boolean {
    const user = this.currentUser();
    if (!user) return false;

    const roleHierarchy: Record<string, number> = { admin: 2, user: 1 };
    const userRoleLevel = roleHierarchy[user.role] || 0;
    const requiredRoleLevel = roleHierarchy[requiredRole] || 0;
    return userRoleLevel >= requiredRoleLevel;
  }

  canReply(): boolean {
    return this.hasPermission('user');
  }

  canToggleBot(): boolean {
    return this.hasPermission('admin');
  }

  validateSession(): Observable<boolean> {
    return this.http.get<User>(`${environment.apiUrl}/v1/auth/me`).pipe(
      tap((user) => {
        if (user) {
          this.setUser(user);
        }
      }),
      map((user) => !!user),
      catchError(() => {
        this.clearSession();
        return of(false);
      })
    );
  }

  private setUser(user: User): void {
    localStorage.setItem(this.USER_KEY, JSON.stringify(user));
    this.currentUser.set(user);
    this.isAuthenticated.set(true);
  }

  private clearSession(): void {
    localStorage.removeItem(this.USER_KEY);
    this.currentUser.set(null);
    this.isAuthenticated.set(false);
  }

  private initSession(): void {
    // First, try to load cached user for quick UI
    const userJson = localStorage.getItem(this.USER_KEY);
    if (userJson) {
      try {
        const cachedUser = JSON.parse(userJson) as User;
        this.currentUser.set(cachedUser);
        this.isAuthenticated.set(true);
      } catch {
        localStorage.removeItem(this.USER_KEY);
      }
    }

    // Validate session with backend (cookie will be sent automatically)
    this.http.get<User>(`${environment.apiUrl}/v1/auth/me`).pipe(
      catchError(() => of(null))
    ).subscribe((user) => {
      if (user) {
        this.setUser(user);
      } else {
        this.clearSession();
      }
      this.isInitialized.set(true);
    });
  }
}
