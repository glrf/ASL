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
      email: 'userEmail',
      password: 'userPassword'
    });
  }
}
