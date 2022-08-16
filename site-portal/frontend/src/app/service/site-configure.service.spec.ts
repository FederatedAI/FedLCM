import { TestBed } from '@angular/core/testing';

import { SiteConfigureService } from './site-configure.service';

describe('SiteConfigureService', () => {
  let service: SiteConfigureService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SiteConfigureService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
