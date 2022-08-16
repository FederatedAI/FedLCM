import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CertificateAuthorityDetailComponent } from './certificate-authority-detail.component';

describe('CertificateDetailComponent', () => {
  let component: CertificateAuthorityDetailComponent;
  let fixture: ComponentFixture<CertificateAuthorityDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CertificateAuthorityDetailComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CertificateAuthorityDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
