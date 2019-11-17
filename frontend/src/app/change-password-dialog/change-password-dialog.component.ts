import {Component, Inject, OnInit} from '@angular/core';
import {FormControl} from '@angular/forms';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material';
import {ChangePasswordDialogData} from '../entities/changePasswordDialogData';

@Component({
  selector: 'app-change-password-dialog',
  templateUrl: './change-password-dialog.component.html',
  styleUrls: ['./change-password-dialog.component.css']
})
export class ChangePasswordDialogComponent implements OnInit {

  newPasswordField = new FormControl('');

  constructor(
    public dialogRef: MatDialogRef<ChangePasswordDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: ChangePasswordDialogData) { }

  ngOnInit() {
  }

  changePassword() {
    // password has to be changed - either here or send back to user UserDetailComponent
    this.data.newPassword = this.newPasswordField.value;
    this.dialogRef.close(this.data);
  }
  cancelAndCloseDialog() {
    this.dialogRef.close(-1);
  }

}
