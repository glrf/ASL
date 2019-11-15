import { Injectable } from '@angular/core';
import {Observable, of} from 'rxjs';
import {User} from './entities/user';
import {HttpClient, HttpParams} from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private baseUrl = '';

  constructor(private http: HttpClient) {
  }

  getUserInfo(userId: string): Observable<User> {
    /*return this.http.get<User>(this.baseUrl + '/user/detail/get', {
      params: new HttpParams().set('uid', userId)
    });*/
    return of({
      uid: userId,
      firstName: 'userFirstName',
      lastName: 'userLastName',
      email: 'user@email.ch',
    });
  }

  saveUserInfo(userInfo: User): Observable<boolean> {
    // return this.http.post<boolean>(this.baseUrl + '/user/detail/update', userInfo);
    return of(true);
  }

  changeUserPassword(userId: string, newPassword: string): Observable<boolean> {
    // return this.http.post<boolean>(this.baseUrl + '/user/password/change', userInfo);
    return of(true);
  }
}
