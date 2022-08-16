import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateOpenflComponent } from './create-openfl-fed.component';

describe('CreateOpenflComponent', () => {
  let component: CreateOpenflComponent;
  let fixture: ComponentFixture<CreateOpenflComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateOpenflComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateOpenflComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
