import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HighJsonComponent } from './high-json.component';

describe('HighJsonComponent', () => {
  let component: HighJsonComponent;
  let fixture: ComponentFixture<HighJsonComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ HighJsonComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(HighJsonComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
