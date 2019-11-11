import { Certificate } from './certificate';
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import {HttpClient, HttpParams} from '@angular/common/http';


@Injectable({
  providedIn: 'root'
})
export class CertificateService {

  private baseUrl = '';

  constructor(private http: HttpClient) { }

  getCertificates(userid: string): Observable<Certificate[]> {
    /*return this.http.get<Certificate[]>(this.baseUrl + '/certificate/get/', {
      params : new HttpParams().set('userid', userid)
    });*/
    return of([
      new Certificate('1392932')
    ]);
  }

  issueCertificate(uid: string): Observable<Certificate> {
    /*return this.http.post<Certificate>(this.baseUrl + '/certificate/issue/', uid);*/
    return of( new Certificate('1834398')
    );
  }

  revokeCertificates(certificates: Certificate[]): Observable<boolean> {
    /*return this.http.post<boolean>(this.baseUrl + '/certificate/revoke/', certificates);*/
    return of(true);
  }
}
