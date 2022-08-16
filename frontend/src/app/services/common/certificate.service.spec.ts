import { TestBed } from '@angular/core/testing';

import { CertificateMgService } from './certificate.service';

describe('CertificateMgService', () => {
  let service: CertificateMgService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CertificateMgService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
