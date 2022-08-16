import { ComponentFixture, TestBed } from '@angular/core/testing';

import { JobNewComponent } from './job-new.component';

describe('JobNewComponent', () => {
  let component: JobNewComponent;
  let fixture: ComponentFixture<JobNewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ JobNewComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(JobNewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
