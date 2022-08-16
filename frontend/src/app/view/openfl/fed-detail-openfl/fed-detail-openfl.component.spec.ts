import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FedDetailOpneFLComponent } from './fed-detail-openfl.component';

describe('FedDetailFateComponent', () => {
  let component: FedDetailOpneFLComponent;
  let fixture: ComponentFixture<FedDetailOpneFLComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FedDetailOpneFLComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(FedDetailOpneFLComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
