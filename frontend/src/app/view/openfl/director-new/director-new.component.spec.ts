import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DirectorNewComponent } from './director-new.component';

describe('ExchangeNewComponent', () => {
  let component: DirectorNewComponent;
  let fixture: ComponentFixture<DirectorNewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DirectorNewComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DirectorNewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
