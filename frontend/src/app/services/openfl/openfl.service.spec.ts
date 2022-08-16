import { TestBed } from '@angular/core/testing';

import { OpenflService } from './openfl.service';

describe('OpenflService', () => {
  let service: OpenflService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(OpenflService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
