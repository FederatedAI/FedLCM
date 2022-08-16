import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FedMgComponent } from './fed-mg.component';

describe('FedMgComponent', () => {
  let component: FedMgComponent;
  let fixture: ComponentFixture<FedMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FedMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(FedMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
