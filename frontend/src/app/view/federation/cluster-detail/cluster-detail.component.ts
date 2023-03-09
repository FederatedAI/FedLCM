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
  selector: 'app-cluster-detail',
  templateUrl: './cluster-detail.component.html',
  styleUrls: ['./cluster-detail.component.scss']
})
export class ClusterDetailComponent implements OnInit {
  isShowDetailFailed = false
  isPageLoading = true
  errorMessage = ''
  //uuid is the federation UUID
  uuid = ''
  cluster_uuid = ''
  deleteType = 'cluster'
  forceRemove = false
  constantGather = constantGather
  clusterDetail: any = {}
  openDeleteModal = false
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  accessInfoList: { [key: string]: any }[] = []
  ingressInfoList: { [key: string]: any }[] = []
  constructor(private route: ActivatedRoute, private router: Router, private fedService: FedService, private chartservice: ChartService) { }
  ngOnInit(): void {
    this.getClusterDetail()
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
      if (this.clusterDetail.deployment_yaml) {
        this.code.setValue(this.clusterDetail.deployment_yaml)
      }
    } else {
      this.code = null
    }
    this.overview = value
  }

  isManagedCluster = true;
  //getClusterDetail is to get the Cluster Detail
  getClusterDetail() {
    this.isShowDetailFailed=false
    this.openDeleteModal = false
    this.isDeleteSubmit = false
    this.isDeleteFailed = false
    this.isShowDetailFailed = false
    this.isPageLoading = true
    this.errorMessage = ''
    this.uuid = ''
    this.cluster_uuid = ''
    this.accessInfoList = []
    this.ingressInfoList = []
    this.route.params.subscribe(
      value => {
        this.uuid = value.id
        this.cluster_uuid = value.cluster_uuid
        if (this.uuid && this.cluster_uuid) {
          this.fedService.getClusterInfo(value.id, value.cluster_uuid).subscribe(
            data => {
              this.clusterDetail = data.data
              this.isManagedCluster = this.clusterDetail?.is_managed
              this.getChartDetail(data.data.chart_uuid)
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
              for (const key in data.data.ingress_info) {
                const obj: any = {
                  name: key,
                }
                const value = data.data.ingress_info[key]
                if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
                  for (const key2 in value) {
                    obj[key2] = value[key2]
                  }
                }
                this.ingressInfoList.push(obj)
              }
              this.isPageLoading = false

            },
            err => {
              this.isShowDetailFailed = true
              this.isPageLoading = false
              if (err.error.message) this.errorMessage = err.error.message
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
    this.getClusterDetail()
  }
  
  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal() {
    this.openDeleteModal = true
  }

  get accessInfo() {
    return JSON.stringify(this.clusterDetail.access_info) === '{}'
  }

  //toSitePortal is to open the SitePortal if the service is successfully deployed
  toSitePortal(item: any) {
    window.open(
      `https://${item.host}:${item.port}`, '_blank'
    )
  }

  //confirmDelete to submit the request of 'Delete cluster'
  confirmDelete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.fedService.deleteParticipant(this.uuid, this.deleteType, this.cluster_uuid, this.forceRemove)
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
