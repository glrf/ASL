import { Component, OnInit } from '@angular/core';
import {CertificateService} from "../certificate.service";
import { Certificate } from '../certificate'

@Component({
  selector: 'app-certificates',
  templateUrl: './certificates.component.html',
  styleUrls: ['./certificates.component.css']
})
export class CertificatesComponent implements OnInit {

  constructor(private certificateService: CertificateService) { }

  certificates: Certificate[];

  ngOnInit() {
    this.getCertificates();
  }

  getCertificates(): void {
    this.certificateService.getCertificates().subscribe(certificates => this.certificates = certificates)
  }

}
