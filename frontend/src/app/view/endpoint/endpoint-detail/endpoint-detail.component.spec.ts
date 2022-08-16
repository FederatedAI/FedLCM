import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EndpointDetailComponent } from './endpoint-detail.component';

describe('EndpointDetailComponent', () => {
  let component: EndpointDetailComponent;
  let fixture: ComponentFixture<EndpointDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ EndpointDetailComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(EndpointDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
