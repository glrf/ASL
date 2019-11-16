import { AuthConfig } from 'angular-oauth2-oidc';

export const authConfig: AuthConfig = {

  // Url of the Identity Provider
  issuer: 'https://hydra.fadalax.tech:9000/',

  // URL of the SPA to redirect the user to after login
  redirectUri: 'https://fadalax.tech/index.html',

  // The SPA's id. The SPA is registered with this id at the auth-server
  clientId: 'fadalax-frontend',

  // set the scope for the permissions the client should request
  // The first three are defined by OIDC. The 4th is a usecase-specific one
  scope: 'openid',

  responseType: 'id_token',
};

