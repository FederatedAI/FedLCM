import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UserMgComponent } from './user-mg.component';

describe('UserMgComponent', () => {
  let component: UserMgComponent;
  let fixture: ComponentFixture<UserMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ UserMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(UserMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
