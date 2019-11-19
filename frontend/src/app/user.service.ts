import { Injectable } from '@angular/core';
import {Observable} from 'rxjs';
import {User} from './entities/user';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import {OAuthService} from 'angular-oauth2-oidc';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private baseUrl = 'https://idp.fadalax.tech/';

  constructor(private http: HttpClient, private oauthService: OAuthService) {
  }

  logOut() {
      this.oauthService.logOut();
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

  issueCertificate(): Observable<string> {
    return this.http.get(this.baseUrl + 'cert', {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken()),
      responseType: 'text'
    }).pipe(
      map(res => {
        return res;
      })
    );
  }


  public downloadResource(): Observable<Blob> {
    return this.http.get(this.baseUrl + 'cert', {
      headers: new HttpHeaders('Authorization: Bearer ' + this.oauthService.getIdToken()),
      responseType: 'blob'});
  }

  revokeCertificates(): Observable<boolean> {
    return this.http.delete(this.baseUrl + 'cert', {
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
