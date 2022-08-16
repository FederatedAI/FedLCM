import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ModelMgComponent } from './model-mg.component';

describe('ModelMgComponent', () => {
  let component: ModelMgComponent;
  let fixture: ComponentFixture<ModelMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ModelMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ModelMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
