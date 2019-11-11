import {Component, Input, OnInit} from '@angular/core';
import {User} from '../entities/user';
import {UserService} from '../user.service';
import {FormControl, Validators} from '@angular/forms';
import {first} from 'rxjs/operators';
import {MatDialog, MatDialogRef} from '@angular/material';
import {ChangePasswordDialogComponent} from '../change-password-dialog/change-password-dialog.component';
import {ChangePasswordDialogData} from '../entities/changePasswordDialogData';

@Component({
  selector: 'app-user-detail',
  templateUrl: './user-detail.component.html',
  styleUrls: ['./user-detail.component.css']
})
export class UserDetailComponent implements OnInit {

  @Input()
  private uid: string;

  private userInfo: User;
  private editEnabled = false;

  firstNameField = new FormControl('');
  lastNameField = new FormControl('');

  emailField = new FormControl('', [Validators.required, Validators.email]);

  getErrorMessage() {
    return this.emailField.hasError('required') ? 'You must enter a value' :
      this.emailField.hasError('email') ? 'Not a valid email' :
        '';
  }

  constructor(private userService: UserService, private dialog: MatDialog) {
  }

  ngOnInit() {
    this.userService.getUserInfo(this.uid)
      .subscribe(user => {
        this.userInfo = user;
        this.firstNameField.setValue(this.userInfo.firstName);
        this.lastNameField.setValue(this.userInfo.lastName);
        this.emailField.setValue(this.userInfo.email);
      });
  }

  enableEditUserInfo() {
    this.editEnabled = true;
  }

  saveUserInfo() {
    const modifiedUser = {
      uid: this.userInfo.uid,
      firstName: this.firstNameField.value,
      lastName: this.lastNameField.value,
      email: this.emailField.value,
    };
    this.userService.saveUserInfo(modifiedUser)
      .subscribe(success => {
        if (success) {
          this.userInfo = modifiedUser;
          this.editEnabled = false;
        }
      });
  }

  startIssueCertificateProcess() {
  }

  startChangePasswordProcess() {
    const dialogRef = this.dialog.open(ChangePasswordDialogComponent, {
      data: {
        newPassword: null
      }
    });
    dialogRef.afterClosed().subscribe(result => {
      if (result !== -1) {
        // user wants to change password - result is of type ChangePasswordDialogData
        const passwordData = (result as ChangePasswordDialogData);
        this.userService.changeUserPassword(this.uid, passwordData.newPassword)
          .subscribe(success => console.log('Password change successfull: ' + success));
      } // else: user cancelled change password request / nothing has to be done
    });
  }

}
