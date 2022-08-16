import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TimeOutServiceComponent } from './time-out-service.component';

describe('TimeOutServiceComponent', () => {
  let component: TimeOutServiceComponent;
  let fixture: ComponentFixture<TimeOutServiceComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TimeOutServiceComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TimeOutServiceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
