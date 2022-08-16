import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InfraDetailComponent } from './infra-detail.component';

describe('InfraDetailComponent', () => {
  let component: InfraDetailComponent;
  let fixture: ComponentFixture<InfraDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ InfraDetailComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(InfraDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
