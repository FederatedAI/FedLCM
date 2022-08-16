import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ExchangeNewComponent } from './exchange-new.component';

describe('ExchangeNewComponent', () => {
  let component: ExchangeNewComponent;
  let fixture: ComponentFixture<ExchangeNewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ExchangeNewComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ExchangeNewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
