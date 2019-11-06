import { Certificate } from './certificate'
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';


@Injectable({
  providedIn: 'root'
})
export class CertificateService {

  constructor() { }

  getCertificates(): Observable<Certificate[]> {
    return of([
      new Certificate(1)
    ]);
  }
}
