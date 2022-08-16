import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ProjectMgComponent } from './project-mg.component';

describe('ProjectMgComponent', () => {
  let component: ProjectMgComponent;
  let fixture: ComponentFixture<ProjectMgComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ProjectMgComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectMgComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
