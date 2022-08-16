import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CertificateMgComponent } from './certificate-mg.component';

describe('CertificateMgComponent', () => {
  let component: CertificateMgComponent;
  let fixture: ComponentFixture<CertificateMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CertificateMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CertificateMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
