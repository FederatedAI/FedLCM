import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FedDetailFateComponent } from './fed-detail-fate.component';

describe('FedDetailFateComponent', () => {
  let component: FedDetailFateComponent;
  let fixture: ComponentFixture<FedDetailFateComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FedDetailFateComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(FedDetailFateComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
