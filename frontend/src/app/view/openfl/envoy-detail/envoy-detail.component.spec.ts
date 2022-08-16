import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EnvoyDetailComponent } from './envoy-detail.component';

describe('ClusterDetailComponent', () => {
  let component: EnvoyDetailComponent;
  let fixture: ComponentFixture<EnvoyDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ EnvoyDetailComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(EnvoyDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
