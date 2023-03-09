// Copyright 2022 VMware, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router'
import { FedService } from 'src/app/services/federation-fate/fed.service'
import { constantGather } from 'src/utils/constant';
import { ChartService } from 'src/app/services/common/chart.service';

@Component({
  selector: 'app-exchange-detail',
  templateUrl: './exchange-detail.component.html',
  styleUrls: ['./exchange-detail.component.scss']
})
export class ExchangeDetailComponent implements OnInit {
  isShowDetailFailed = false
  isPageLoading = true
  errorMessage = ''
  //uuid is the federation UUID
  uuid = ''
  exchange_uuid = ''
  constantGather = constantGather
  exchangeDetail: any = {}
  openDeleteModal = false
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  accessInfoList: { [key: string]: any }[] = []
  constructor(private route: ActivatedRoute, private router: Router, private fedService: FedService, private chartservice: ChartService) { }
  ngOnInit(): void {
    this.getExchangeDetail()
  }
  code: any
  overview = true
  get isOverview() {
    return this.overview
  }
  set isOverview(value) {
    if (value) {
      const yamlHTML = document.getElementById('yaml') as any
      this.code = window.CodeMirror.fromTextArea(yamlHTML, {
        value: '',
        mode: 'yaml',
        lineNumbers: true,
        indentUnit: 1,
        lineWrapping: true,
        tabSize: 2,
        readOnly: true
      })
      if (this.exchangeDetail.deployment_yaml) {
        this.code.setValue(this.exchangeDetail.deployment_yaml)
      }
    } else {
      this.code = null
    }
    this.overview = value
  }

  isManagedExchange = true;
  //getExchangeDetail is to get the Exchange Detail
  getExchangeDetail() {
    this.openDeleteModal = false
    this.isDeleteSubmit = false
    this.isDeleteFailed = false
    this.isShowDetailFailed = false
    this.isPageLoading = true
    this.errorMessage = ''
    this.uuid = ''
    this.exchange_uuid = ''
    this.route.params.subscribe(
      value => {
        this.uuid = value.id
        this.exchange_uuid = value.exchange_uuid
        if (this.uuid && this.exchange_uuid) {
          this.fedService.getExchangeInfo(value.id, value.exchange_uuid).subscribe(
            data => {
              this.exchangeDetail = data.data
              this.isManagedExchange = this.exchangeDetail?.is_managed
              if (this.isManagedExchange) this.getChartDetail(data.data.chart_uuid)
              const value = data.data.deployment_yaml
              if (this.code) {
                this.code.setValue(value)
              }
              for (const key in data.data.access_info) {
                const obj: any = {
                  name: key,
                }
                const value = data.data.access_info[key]
                if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
                  for (const key2 in value) {
                    obj[key2] = value[key2]
                  }
                }
                this.accessInfoList.push(obj)
              }
              this.isPageLoading = false
            },
            err => {
              this.isPageLoading = false
              if (err.error.message) this.errorMessage = err.error.message
              this.isShowDetailFailed = true
            }
          )
        }
      }
    )

  }

  isChartContainsPortalservices = false;
  isShowChartFailed = false;
  //getChartDetail is to get the chart if 'isChartContainsPortalservices'
  getChartDetail(uuid: string) {
    this.chartservice.getChartDetail(uuid)
      .subscribe((data: any) => {
        if (data.data) {
          //isChartContainsPortalservices to decide the certificate we need when creating the exchange
          this.isChartContainsPortalservices = data.data.contain_portal_services;
        }
        this.isPageLoading = false;
        this.isShowChartFailed = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
  }

  //refresh is for refresh button
  refresh() {
    this.accessInfoList = []
    this.getExchangeDetail()
  }
  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal() {
    this.openDeleteModal = true
  }

  get accessInfo() {
    return JSON.stringify(this.exchangeDetail.access_info) === '{}'
  }
  
  deleteType = 'exchange'
  forceRemove = false
  //confirmDelete to submit the request of 'Delete exchange'
  confirmDelete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.fedService.deleteParticipant(this.uuid, this.deleteType, this.exchange_uuid, this.forceRemove)
      .subscribe(() => {
        this.router.navigate(['/federation/fate', this.uuid]);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });

  }

  // Exchange and cluster jump to the upgrade page through toUpgrade
  toUpgrade(item: {uuid: string, name: string, version: string}, type: string) {
    this.router.navigate(['/federation', 'fate', this.uuid, 'detail', item.uuid, item.version, type+'-'+item.name, 'upgrade'])
  }

  back() {
    this.router.navigate(['federation', 'fate', this.uuid])
  }
}
