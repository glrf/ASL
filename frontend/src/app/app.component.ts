import {Component} from '@angular/core';
import {NullValidationHandler, OAuthService} from 'angular-oauth2-oidc';
import {authConfig} from './auth.config';
import {HttpClient} from '@angular/common/http';
import {error} from 'util';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'ASL';

  validationUrl = 'https://hydra.fadalax.tech:9000';

  constructor(private oauthService: OAuthService, private http: HttpClient) {
    this.configure();
  }

  private configure() {
    this.oauthService.configure(authConfig);
    this.oauthService.events.subscribe(event => console.log(event));
    this.oauthService.tokenValidationHandler = new NullValidationHandler();
    this.oauthService.loadDiscoveryDocumentAndTryLogin({
      onTokenReceived: context => {
        console.log('logged in');
        console.log('User:' + this.oauthService.getIdentityClaims()['sub']);
      },
      onLoginError: context => console.log(context),

    }).then(_ => {
      if (!this.oauthService.hasValidIdToken()) {
        this.oauthService.initImplicitFlow();
        return false;
      } else {
        return true;
      }
    });
  }
}
