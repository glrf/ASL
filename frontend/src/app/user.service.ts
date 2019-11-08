import { Injectable } from '@angular/core';
import {Observable, of} from 'rxjs';
import {User} from './entities/user';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  constructor() {
  }

  getUserInfo(userId: string): Observable<User> {
    return of({
      uid: userId,
      firstName: 'userFirstName',
      lastName: 'userLastName',
      email: 'user@email.ch',
      password: 'userPassword'
    });
  }

  saveUserDetailChanges(userInfo: User): Observable<boolean> {
    return of(true);
  }

  changePassword(userId: string, newPassword: string): Observable<boolean> {
    return of(true);
  }


}
