import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartDetailComponent } from './chart-detail.component';

describe('ChartDetailComponent', () => {
  let component: ChartDetailComponent;
  let fixture: ComponentFixture<ChartDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChartDetailComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ChartDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
