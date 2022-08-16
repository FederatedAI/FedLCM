import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DataMgComponent } from './data-mg.component';

describe('DataMgComponent', () => {
  let component: DataMgComponent;
  let fixture: ComponentFixture<DataMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DataMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DataMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
