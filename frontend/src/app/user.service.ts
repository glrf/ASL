import { Injectable } from '@angular/core';
import {Observable, of} from 'rxjs';
import {User} from './entities/user';
import {HttpClient, HttpHeaders, HttpParams} from '@angular/common/http';
import {OAuthService} from 'angular-oauth2-oidc';
import {tokenize} from '@angular/compiler/src/ml_parser/lexer';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private baseUrl = 'https://idp.fadalax.tech/';

  constructor(private http: HttpClient, private oauthService: OAuthService) {
  }

  getUserInfo(): Observable<User> {
    this.oauthService.events.subscribe()
    return this.http.get<User>(this.baseUrl + 'user', {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken())
    });
  }

  saveUserInfo(userInfo: User): Observable<boolean> {
    // return this.http.post<boolean>(this.baseUrl + '/user', userInfo);
    return of(true);
  }

  changeUserPassword(userId: string, newPassword: string): Observable<boolean> {
    // return this.http.post<boolean>(this.baseUrl + '/user/password/change', newPassword);
    return of(true);
  }
}
