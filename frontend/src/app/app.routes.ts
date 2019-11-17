import {Routes} from '@angular/router';
import {HomeComponent} from './home/home.component';
import {AuthGuard} from './auth.guard';


export const APP_ROUTES: Routes = [
  {
    path: '**',
    component: HomeComponent,
    canActivate: [AuthGuard]
  }
];

