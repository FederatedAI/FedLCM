import { TestBed } from '@angular/core/testing';

import { UserMgService } from './user-mg.service';

describe('UserMgService', () => {
  let service: UserMgService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(UserMgService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
