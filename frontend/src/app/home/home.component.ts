import { Component, OnInit } from '@angular/core';
import {OAuthService} from 'angular-oauth2-oidc';
import {UserService} from '../user.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  uid = null;

  constructor(private oauthService: OAuthService) {
  }

  ngOnInit() {
    const claims = this.oauthService.getIdentityClaims();
    this.uid = claims['sub'];
    console.log(this.uid);
  }
}
