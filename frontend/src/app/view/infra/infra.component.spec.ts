import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InfraComponent } from './infra.component';

describe('InfraComponent', () => {
  let component: InfraComponent;
  let fixture: ComponentFixture<InfraComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ InfraComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(InfraComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
