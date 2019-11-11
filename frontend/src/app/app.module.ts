import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CertificatesComponent } from './certificates/certificates.component';
import {HttpClientModule} from '@angular/common/http';
import {MatButtonModule} from '@angular/material';
import { CertificateDetailComponent } from './certificate-detail/certificate-detail.component';
import { CertificateAdminViewComponent } from './certificate-admin-view/certificate-admin-view.component';

@NgModule({
  declarations: [
    AppComponent,
    CertificatesComponent,
    CertificateDetailComponent,
    CertificateAdminViewComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    MatButtonModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
