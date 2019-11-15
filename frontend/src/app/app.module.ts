import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { UserDetailComponent } from './user-detail/user-detail.component';
import {MatButtonModule, MatDialogModule, MatFormFieldModule, MatInputModule} from '@angular/material';
import {ReactiveFormsModule} from '@angular/forms';
import { ChangePasswordDialogComponent } from './change-password-dialog/change-password-dialog.component';
import {HttpClientModule} from '@angular/common/http';
import {OAuthModule} from 'angular-oauth2-oidc';
import { HomeComponent } from './home/home.component';
import {RouterModule} from '@angular/router';
import {User} from './entities/user';

@NgModule({
  declarations: [
    AppComponent,
    UserDetailComponent,
    ChangePasswordDialogComponent,
    HomeComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    MatButtonModule,
    MatFormFieldModule,
    ReactiveFormsModule,
    MatInputModule,
    MatDialogModule,
    HttpClientModule,
    OAuthModule.forRoot(),
    RouterModule.forRoot([
      {path: '', component: HomeComponent},
      {path: 'token', component: UserDetailComponent}
    ])
  ],
  entryComponents: [
    ChangePasswordDialogComponent
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
