import { TestBed } from '@angular/core/testing';

import { InfraService } from './infra.service';

describe('InfraService', () => {
  let service: InfraService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(InfraService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
