import { TestBed } from '@angular/core/testing';

import { FedService } from './fed.service';

describe('FedService', () => {
  let service: FedService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(FedService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
