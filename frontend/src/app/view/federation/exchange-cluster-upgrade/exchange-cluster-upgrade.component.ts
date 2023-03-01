import { Component, OnInit } from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { FedService } from 'src/app/services/federation-fate/fed.service';

@Component({
  selector: 'app-exchange-cluster-upgrade',
  templateUrl: './exchange-cluster-upgrade.component.html',
  styleUrls: ['./exchange-cluster-upgrade.component.scss']
})
export class ExchangeClusterUpgradeComponent implements OnInit {
  form!: FormGroup;
  upgradeVersionList:string[] = []
  isShowChartFailed = false
  errorMessage = ''
  title = ''
  fedUuid = ''
  upgradeUuid = ''
  type!: 'cluster' | 'exchange'
  constructor(private formBuilder: FormBuilder, private fedservice: FedService, private route: ActivatedRoute, private router: Router) { 
    this.form = this.formBuilder.group({
      version: this.formBuilder.group({
        version: ['']
      })
    })
  }

  ngOnInit(): void {
    this.route.params.subscribe(
      value => {
        this.title  = value.name
        this.fedUuid  = value.id
        this.upgradeUuid  = value.uuid
        this.type = this.title.split('-')[0].toLocaleLowerCase() as 'cluster' | 'exchange'
        this.getExchangeClusterUpgradeVersionList(this.fedUuid, this.upgradeUuid, this.type)
      }
    )
  }

  getExchangeClusterUpgradeVersionList(fed_uuid: string, upgrade_uuid: string, type: 'cluster' | 'exchange') {
    this.fedservice.getExchangeClusterUpgradeVersionList(fed_uuid, upgrade_uuid, type).subscribe(
      data => {
        this.upgradeVersionList = data.data.upgradeable_version_list
      }
    )
  }

  upgradeExchangeCluster(fed_uuid: string, upgrade_uuid: string, type: 'cluster' | 'exchange') {
    this.fedservice.upgradeExchangeCluster(fed_uuid, upgrade_uuid, type, {upgradeVersion: this.form.controls['version'].get('version')?.value}).subscribe(
      data => {
        this.router.navigate(['federation', 'fate', this.fedUuid, this.type, 'detail', this.upgradeUuid])
      },
      err => {
        this.errorMessage = err.error.message;
        this.isShowChartFailed = true
      }
    )
  }
}
