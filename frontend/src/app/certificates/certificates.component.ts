import {Component, OnInit} from '@angular/core';
import {CertificateService} from '../certificate.service';
import {Certificate} from '../certificate';

@Component({
  selector: 'app-certificates',
  templateUrl: './certificates.component.html',
  styleUrls: ['./certificates.component.css']
})
export class CertificatesComponent implements OnInit {

  constructor(private certificateService: CertificateService) {
  }

  private uid = '';
  private certificates: Certificate[] = [];
  private selectedCertificate = null;

  // Admin view info
  private numIssuedCertificates = 0;
  private numRevokedCertificates = 0;
  private lastCertificateSerialNumber = '-';

  ngOnInit() {
    this.getCertificates();
  }

  getCertificates(): void {
    this.certificateService.getCertificates(this.uid).subscribe(certificates => this.certificates = certificates);
  }

  displayCertificateDetails(c: Certificate) {
    this.selectedCertificate = c;
  }

  downloadCertificate(certificate: Certificate) {
    // TODO: download certificate
  }

  issueCertificate() {
    this.certificateService.issueCertificate(this.uid).subscribe(result => {
      this.certificates = this.certificates.concat(result);
      this.numIssuedCertificates += 1;
      this.lastCertificateSerialNumber = result.serialNumber;
    });
  }

  revokeCertificates() {
    this.certificateService.revokeCertificates(this.certificates).subscribe(success => {
      if (success) {
        this.numRevokedCertificates = this.numRevokedCertificates + this.certificates.length;
        this.certificates = [];
        this.selectedCertificate = null;
      }
    });
  }
}
