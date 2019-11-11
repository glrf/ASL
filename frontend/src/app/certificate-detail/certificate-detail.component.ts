import {Component, Inject, Input, OnInit} from '@angular/core';
import {Certificate} from '../certificate';

@Component({
  selector: 'app-certificate-detail',
  templateUrl: './certificate-detail.component.html',
  styleUrls: ['./certificate-detail.component.css']
})
export class CertificateDetailComponent implements OnInit {

  @Input()
  certificate: Certificate;

  constructor() { }

  ngOnInit() {
  }

  downloadCertificate(certificate: Certificate) {
    // TODO: download certificate
  }

}
