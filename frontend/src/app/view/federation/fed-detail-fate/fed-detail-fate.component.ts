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
import { ActivatedRoute, Router } from '@angular/router';
import { FedService } from 'src/app/services/federation-fate/fed.service';
import { ParticipantFATEStatus, ParticipantFATEType, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-fed-detail-fate',
  templateUrl: './fed-detail-fate.component.html',
  styleUrls: ['./fed-detail-fate.component.scss']
})
export class FedDetailFateComponent implements OnInit {

  constructor(private fedservice: FedService, private router: Router, private route: ActivatedRoute) {
    this.showFedDetail(this.uuid);
    this.showParticipantList(this.uuid)
  }

  ngOnInit(): void {
  }

  participantFATEstatus = ParticipantFATEType;
  participantFATEtype = ParticipantFATEStatus;
  constantGather = constantGather;
  exchangeInfoList: any[] = []
  //uuid is the uuid of current federation
  uuid = String(this.route.snapshot.paramMap.get('id'));

  fedDetail: any;
  errorMessage = "Service Error!"
  isShowFedDetailFailed: boolean = false;
  isPageLoading: boolean = true;
  //showFedDetail is to get the federation detail
  showFedDetail(uuid: string) {
    this.isPageLoading = true;
    this.isShowFedDetailFailed = false;
    this.fedservice.getFedDetail(uuid)
      .subscribe((data: any) => {
        this.fedDetail = data.data;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isPageLoading = false
          this.isShowFedDetailFailed = true
        }
      );
  }

  participantList: any;
  clusterlist: any[] = [];
  exchange: any;
  isShowParticipantListFailed: boolean = false;
  //showParticipantList is to get the list exchange and clusters
  showParticipantList(uuid: string) {
    this.isPageLoading = true;
    this.isShowParticipantListFailed = false;
    this.fedservice.getFedParticipantList(uuid)
      .subscribe((data: any) => {
        this.participantList = data.data;
        this.clusterlist = this.participantList.clusters || [];
        this.exchange = this.participantList.exchange;
        if (this.exchange) {
          for (const key in this.exchange.access_info) {
            const obj: any = {
              name: key,
            }
            const value = this.exchange.access_info[key]
            if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
              for (const key2 in value) {
                obj[key2] = value[key2]
              }
            }
            this.exchangeInfoList.push(obj)
          }
          this.clusterlist.forEach(cluster => {
            cluster.clusterList = []
            for (const key in cluster.access_info) {
              const obj: any = {
                name: key,
              }
              const value = cluster.access_info[key]
              if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
                for (const key2 in value) {
                  obj[key2] = value[key2]
                }
              }
              cluster.clusterList.push(obj)
            }
          });
        }
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isPageLoading = false
          this.isShowParticipantListFailed = true
        }
      );
  }

  //hasAccessInfo return if the participant contains access info
  hasAccessInfo(object: any): boolean {
    return JSON.stringify(object) != '{}'
  }

  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataList: string = '';
  deleteType: string = '';
  deleteUUID: string = '';
  forceRemove = false;
  //openDeleteConfrimModal is to initial variables when open the modal of 'Delete Certificate'
  openDeleteConfrimModal(type: string, item_uuid: string) {
    this.deleteType = type;
    this.deleteUUID = item_uuid;
    this.isDeleteFailed = false;
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
    this.forceRemove = false;
  }

  //delete() is to delete federation or participant based on the buttom is clicked
  delete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    //delete federation
    if (this.deleteType === 'federation') {
      this.fedservice.deleteFed(this.uuid)
        .subscribe(() => {
          this.router.navigate(['/federation']);
        },
          err => {
            this.isDeleteFailed = true;
            this.errorMessage = err.error.message;
          });
    } else {
      // delete particiapant
      this.fedservice.deleteParticipant(this.uuid, this.deleteType, this.deleteUUID, this.forceRemove)
        .subscribe(() => {
          this.reloadCurrentRoute()
        },
          err => {
            this.isDeleteFailed = true;
            this.errorMessage = err.error.message;
          });
    }
  }

  //refresh is for refresh button
  refresh() {
    this.showFedDetail(this.uuid);
    this.showParticipantList(this.uuid)
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  //toDetail is redirect to the participant detail
  toDetail(type: string, detailId: string, info: any) {
    this.route.params.subscribe(
      value => {
        this.router.navigateByUrl(`federation/fate/${value.id}/${type}/detail/${detailId}`)
      }
    )
  }

  //createClusterDisabled is to disabled the 'new cluster' button when there is no active exchange in the current federation
  get createClusterDisabled() {
    if (this.exchange && this.exchange.status === 1) {
      return false
    }
    return true
  }

  // Exchange and cluster jump to the upgrade page through toUpgrade
  toUpgrade(item: {uuid: string, name: string}, type: string) {
    this.router.navigate(['/federation', 'fate', this.uuid, 'detail', item.uuid, type+'-'+item.name, 'upgrade'])
  }

}
