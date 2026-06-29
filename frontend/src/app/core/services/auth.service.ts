import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { tap, catchError, switchMap, filter, take } from 'rxjs/operators';
import { Router } from '@angular/router';
import {AuthRequest, UserResponse} from '../../models/auth.model';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http   = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly API = '/api/v1/auth';

  private readonly loggedIn = new BehaviorSubject<boolean | null>(null);
  isLoggedIn$ = this.loggedIn.asObservable();

  private refreshing = false;
  private readonly refreshSubject = new BehaviorSubject<boolean | null>(null);

  register(req: AuthRequest): Observable<UserResponse> {
    return this.http.post<UserResponse>(`${this.API}/register`, req).pipe(
      catchError(err => throwError(() => err))
    );
  }

  login(req: AuthRequest): Observable<void> {
    return this.http.post<void>(`${this.API}/login`, req).pipe(
      tap(() => this.loggedIn.next(true)),
      catchError(err => throwError(() => err))
    );
  }

  logout(): void {
    this.http.post<void>(`${this.API}/logout`, {}).pipe(
      catchError(() => throwError(() => null))
    ).subscribe({
      complete: () => this.clearSession(),
      error:    () => this.clearSession(),
    });
  }

  checkSession(): Observable<UserResponse> {
    return this.http.get<UserResponse>(`${this.API}/me`).pipe(
      tap(() => this.loggedIn.next(true)),
      catchError(err => {
        this.loggedIn.next(false);
        return throwError(() => err);
      })
    );
  }

  refresh(): Observable<void> {
    if (this.refreshing) {
      return this.refreshSubject.pipe(
        filter(result => result !== null),
        take(1),
        switchMap(success => success
          ? new Observable<void>(obs => { obs.next(); obs.complete(); })
          : throwError(() => new Error('Refresh failed'))
        )
      );
    }

    this.refreshing = true;
    this.refreshSubject.next(null);

    return this.http.post<void>(`${this.API}/refresh`, {}).pipe(
      tap(() => {
        this.refreshing = false;
        this.refreshSubject.next(true);
      }),
      catchError(err => {
        this.refreshing = false;
        this.refreshSubject.next(false);
        this.clearSession();
        return throwError(() => err);
      })
    );
  }

  private clearSession() {
    this.loggedIn.next(false);
    void this.router.navigate(['/auth']);
  }
}
