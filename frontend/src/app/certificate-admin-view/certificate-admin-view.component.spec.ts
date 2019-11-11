import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { CertificateAdminViewComponent } from './certificate-admin-view.component';

describe('CertificateAdminViewComponent', () => {
  let component: CertificateAdminViewComponent;
  let fixture: ComponentFixture<CertificateAdminViewComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ CertificateAdminViewComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CertificateAdminViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
