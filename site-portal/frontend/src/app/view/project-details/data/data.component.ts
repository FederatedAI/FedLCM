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
import { CustomComparator } from 'src/utils/comparator';
import { ProjectService } from '../../../service/project.service';

export interface AssociateData {
  updated_time: string,
  created_time: string,
  data_provider: string,
  name: string,
  associate_status: boolean
}
export interface DataListResponse {
  code: number,
  data: DataDetail[],
  message: "success"
}
export interface DataDetail {
  creation_time: string,
  data_id: string,
  is_local: boolean,
  name: string,
  providing_site_name: string,
  providing_site_party_id: number,
  providing_site_uuid: string,
  update_time: string
}
export interface LocalData {
  creation_time: string,
  data_id: string,
  name: string,
  selected: boolean
}

@Component({
  selector: 'app-data',
  templateUrl: './data.component.html',
  styleUrls: ['./data.component.css']
})

export class DataComponent implements OnInit {
  options: boolean = false;
  openModal: boolean = false;
  errorMessage: string = "";
  routeParams = this.route.parent!.snapshot.paramMap;
  //uuid is project uuid
  uuid = String(this.routeParams.get('id'));
  associatedDataListResponse: any;
  associatedDataList: any;

  constructor(private route: ActivatedRoute, private projectservice: ProjectService, private router: Router) {
    this.showAssociatedDataList();
  }

  ngOnInit(): void {
    this.showAssociatedDataList();
  }

  isPageLoading: boolean = true;
  showAssociatedDataListFailed: boolean = false;
  // showAssociatedDataList is to get associate data list
  showAssociatedDataList() {
    this.isPageLoading = true;
    this.showAssociatedDataListFailed = false;
    this.projectservice.getAssociatedDataList(this.uuid)
      .subscribe(data => {
        this.associatedDataListResponse = data;
        this.associatedDataList = this.associatedDataListResponse.data?.sort((n1: any, n2: any) => { return this.createTimeComparator.compare(n1, n2) });
        this.isPageLoading = false;
      },
        err => {
          this.showAssociatedDataListFailed = true;
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
        });
  }

  localDataList: any;
  localDataListResponse: any;
  newLocalDataList: LocalData[] = [];
  showLocalDataListFailed: boolean = false;
  // showlocalDataList
  showlocalDataList() {
    this.openModal = true;
    this.newLocalDataList = [];
    this.projectservice.getLocalDataList(this.uuid)
      .subscribe(data => {
        this.localDataListResponse = data;
        this.localDataList = this.localDataListResponse.data;
        for (let data of this.localDataList) {
          const localdata: LocalData = {
            creation_time: "",
            data_id: "",
            name: "",
            selected: false
          };
          localdata.creation_time = data.creation_time;
          localdata.name = data.name;
          localdata.data_id = data.data_id;
          this.newLocalDataList.push(localdata);
        }
      },
        err => {
          this.showLocalDataListFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }


  associateLocalDataSubmit: boolean = false;
  associateLocalDataFailed: boolean = false;
  option: any;
  noSelected: boolean = true;
  //associateLocalData is to associate local data in current project
  associateLocalData() {
    this.associateLocalDataSubmit = true;
    for (let data of this.newLocalDataList) {
      if (data.selected) {
        this.noSelected = false;
        this.projectservice.associateData(this.uuid, data.data_id, data.name)
          .subscribe(data => {
            this.reloadCurrentRoute();
          },
            err => {
              this.associateLocalDataFailed = true;
              this.associateLocalDataSubmit = false;
              this.errorMessage = err.error.message;
            }
          );
      }
    }
    //validation and alert
    if (this.noSelected) {
      this.errorMessage = "Please select at least one.";
      this.associateLocalDataFailed = true;
    }
  }

  submitDeleteFailed: boolean = false;
  deleteDataSubmit: boolean = false;
  //deleteAssociatedLocalData is to send the 'cancel association' request of the data the already in the project
  deleteAssociatedLocalData(data_uuid: string) {
    this.deleteDataSubmit = true;
    this.submitDeleteFailed = false;
    this.projectservice.deleteAssociatedData(this.uuid, data_uuid)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.submitDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }
  
  // reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  // Comparator for data grid
  createTimeComparator = new CustomComparator("creation_time", "string");
  updateTimeComparator = new CustomComparator("update_time", "string");
  partyComparator = new CustomComparator("providing_site_name", "string");
  
  // refresh button
  refresh() {
    this.showAssociatedDataList();
  }
}
