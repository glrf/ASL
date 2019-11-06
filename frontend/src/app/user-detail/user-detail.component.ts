import {Component, Input, OnInit} from '@angular/core';
import {User} from '../entities/user';
import {UserService} from '../user.service';

@Component({
  selector: 'app-user-detail',
  templateUrl: './user-detail.component.html',
  styleUrls: ['./user-detail.component.css']
})
export class UserDetailComponent implements OnInit {

  @Input()
  private uid: string;

  private userInfo: User;


  constructor(private userService: UserService) { }

  ngOnInit() {
    this.userService.getUserInfo(this.uid).subscribe(user => this.userInfo = user);
  }

  startIssueCertificateProcess() {
  }
  startChangePasswordProcess() {
  }

}
