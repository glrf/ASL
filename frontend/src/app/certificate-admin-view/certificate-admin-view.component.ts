import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-certificate-admin-view',
  templateUrl: './certificate-admin-view.component.html',
  styleUrls: ['./certificate-admin-view.component.css']
})
export class CertificateAdminViewComponent implements OnInit {

  @Input()
  private numIssuedCerts;

  @Input()
  private numRevokedCerts;

  @Input()
  private lastCertSerialNumber;

  constructor() { }

  ngOnInit() {
  }

}
