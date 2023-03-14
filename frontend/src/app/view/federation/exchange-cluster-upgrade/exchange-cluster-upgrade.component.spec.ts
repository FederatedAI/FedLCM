import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ExchangeClusterUpgradeComponent } from './exchange-cluster-upgrade.component';

describe('ExchangeClusterUpgradeComponent', () => {
  let component: ExchangeClusterUpgradeComponent;
  let fixture: ComponentFixture<ExchangeClusterUpgradeComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ExchangeClusterUpgradeComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ExchangeClusterUpgradeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
