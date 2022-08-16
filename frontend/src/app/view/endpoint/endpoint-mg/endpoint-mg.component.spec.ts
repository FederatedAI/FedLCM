import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EndpointMgComponent } from './endpoint-mg.component';

describe('EndpointMgComponent', () => {
  let component: EndpointMgComponent;
  let fixture: ComponentFixture<EndpointMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ EndpointMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(EndpointMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
