import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartMgComponent } from './chart-mg.component';

describe('ChartMgComponent', () => {
  let component: ChartMgComponent;
  let fixture: ComponentFixture<ChartMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChartMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ChartMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
