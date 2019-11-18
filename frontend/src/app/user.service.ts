import { Injectable } from '@angular/core';
import {Observable, of} from 'rxjs';
import {User} from './entities/user';
import {HttpClient, HttpHeaders, HttpParams, HttpResponse} from '@angular/common/http';
import {OAuthService} from 'angular-oauth2-oidc';
import {tokenize} from '@angular/compiler/src/ml_parser/lexer';
import {catchError, map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private baseUrl = 'https://idp.fadalax.tech/';

  constructor(private http: HttpClient, private oauthService: OAuthService) {
  }

  getUserInfo(): Observable<User> {
    return this.http.get<User>(this.baseUrl + 'user', {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken()),
    });
  }

  saveUserInfo(userInfo: User): Observable<boolean> {
    return this.http.put(this.baseUrl + 'user', userInfo, {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken()),
      responseType: 'text',
      observe: 'response'
    }).pipe(
      map(response => {
        return response.ok;
      })
    );
  }

  changeUserPassword(userId: string, newPassword: string): Observable<boolean> {
    return this.http.put(this.baseUrl + 'user/password', {password: newPassword}, {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken()),
      responseType: 'text',
      observe: 'response'
    }).pipe(
      map(response => {
        return response.ok;
      })
    );
  }
}
