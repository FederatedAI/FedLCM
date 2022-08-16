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

import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { DataService } from '../../service/data.service';
import * as fileSaver from 'file-saver';
import { ClarityIcons, downloadIcon, trashIcon } from '@cds/core/icon';
import { MessageService } from '../../components/message/message.service'

ClarityIcons.addIcons(trashIcon, downloadIcon);
export interface DataDetailResponse {
  Code: number;
  Message: string;
  Data: DataDetail;
}
export interface DataDetail {
  creation_time: string,
  data_id: string,
  description: string,
  feature_size: number,
  features_array: [],
  filename: string,
  id_meta_info: Id_Meta,
  name: string,
  preview_array: string,
  sample_size: number,
  table_name: string,
  upload_job_status: string
}
export interface Id_Meta {
  id_encryption_type: number,
  id_type: number
}

@Component({
  selector: 'app-data-details',
  templateUrl: './data-details.component.html',
  styleUrls: ['./data-details.component.css']
})
export class DataDetailsComponent implements OnInit, OnDestroy {

  constructor(private route: ActivatedRoute, private dataservice: DataService,
    private router: Router, private msg: MessageService) {
    this.showDataDetail();
  }
  ngOnDestroy(): void {
    this.msg.close()
  }
  openModal: boolean = false;

  ngOnInit(): void {
    this.showDataDetail();
  }

  datadetail: DataDetail = {
    creation_time: '',
    data_id: '',
    description: '',
    feature_size: 0,
    features_array: [],
    filename: '',
    id_meta_info: {
      id_encryption_type: 0,
      id_type: 0
    },
    name: '',
    preview_array: '',
    sample_size: 0,
    table_name: '',
    upload_job_status: ''
  };

  dataDetailResponse: any;
  dataDetailList: DataDetail[] = [];
  errorMessage: string = "";
  featureDimensions: string = "";
  previewArray: any;
  key: any;
  isPageLoading: boolean = true;
  isShowDataDetailFailed: boolean = false;
  //showDataDetail is to get the Data Detail
  showDataDetail() {
    const routeParams = this.route.snapshot.paramMap;
    const productIdFromRoute = String(routeParams.get('data_id'));
    this.dataservice.getDataDetail(productIdFromRoute)
      .subscribe((data: DataDetailResponse) => {
        this.dataDetailResponse = data;
        this.datadetail = this.dataDetailResponse.data;
        this.displayMeta();
        this.featureDimensions = this.datadetail.features_array.toString();
        this.previewArray = JSON.parse(this.datadetail.preview_array);
        this.isPageLoading = false;
        this.key = [];
        for (let keyvalue in this.previewArray[0]) {
          this.key.push(keyvalue);
        }
      },
        err => {
          this.isPageLoading = false;
          this.isShowDataDetailFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  //deleteData is to request for deleting data
  deleteData(data_id: string) {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.dataservice.deleteData(data_id)
      .subscribe(() => {
        this.router.navigate(['/data-management']);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  isDownloadSubmit: boolean = false;
  isDownloadFailed: boolean = false;
  //downloadData is to request for downloading data 
  downloadData(data_id: string, data_name: string) {
    this.isDownloadSubmit = true;
    this.isDownloadFailed = false;
    this.dataservice.downloadDataDetail(data_id)
      .subscribe((data: any) => {
        this.msg.success('serverMessage.download200', 1000)
        this.isDownloadSubmit = true;
        let blob: any = new Blob([data], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        fileSaver.saveAs(blob, data_name);
      }), (error: any) => {
        this.isDownloadFailed = true;
        this.errorMessage = 'Error downloading the file';
      }
      , () => console.info('Data downloaded successfully');
  };

  metaOnChange: boolean = false;
  toggleOnChange: boolean = false;
  options: boolean = false;
  option1: boolean = false;
  option2: boolean = false;
  option3: boolean = false;
  option5: boolean = false;
  option6: boolean = false;
  option4: string = "";
  option7: string = "";
  selectoption1: string = "option3";
  selectoption2: string = "option5";
  //selectionOnChange is triggered when the selection of 'ID Metadata' is changed.
  selectionOnChange(val: any) {
    this.metaOnChange = true;
    this.toggleOnChange = !this.toggleOnChange;
    if (val == "option2") {
      this.option1 = false;
      this.option2 = true;
      this.option3 = false;
    }
    if (val == "option1") {
      this.option1 = true;
      this.option2 = false;
      this.option3 = false;
    }
    if (val == "option3") {
      this.option1 = false;
      this.option2 = false;
      this.option3 = true;
    }
    if (val == "option5") {
      this.option6 = false;
      this.option5 = true;
    }
    if (val == "option6") {
      this.option5 = false;
      this.option6 = true;
    }
  }

  //resetMeta is to reset the selection of ID Metadata
  resetMeta() {
    this.displayMeta();
    this.metaOnChange = false;
    this.toggleOnChange = false;
    this.stopPut = false;
    this.errorpop = false;
  }

  //displayMeta() is to get current selection ID metadata
  displayMeta() {
    if (this.datadetail.id_meta_info === null) {
      this.options = false;
      return;
    }
    this.options = true;
    this.option2 = false;
    this.option6 = false;
    if (this.datadetail.id_meta_info.id_type === 0) this.selectoption1 = "option3";
    if (this.datadetail.id_meta_info.id_type === 1) this.selectoption1 = "option1";
    if (this.datadetail.id_meta_info.id_type === 2) {
      this.selectoption1 = "option2";
      this.option2 = true;
      this.option4 = "IMEI";
    }
    if (this.datadetail.id_meta_info.id_type === 3) {
      this.selectoption1 = "option2";
      this.option2 = true;
      this.option4 = "IDFA";
    }
    if (this.datadetail.id_meta_info.id_type === 4) {
      this.selectoption1 = "option2";
      this.option2 = true;
      this.option4 = "IDFV";
    }
    if (this.datadetail.id_meta_info.id_encryption_type === 0) this.selectoption2 = "option5";
    if (this.datadetail.id_meta_info.id_encryption_type === 1) {
      this.selectoption2 === "option6";
      this.option6 = true;
      this.option7 = "MD5";
    }
    if (this.datadetail.id_meta_info.id_encryption_type === 2) {
      this.selectoption2 === "option6";
      this.option6 = true;
      this.option7 = "SHA25";
    }
  }

  stopPut: boolean = false;
  errorpop: boolean = false;
  selecterrorMessage: string = "";
  //checkMeta() is to validate the selection before save
  checkMeta() {
    this.stopPut = false;
    //if ID Metadata is not enabled
    if (!this.options) {
      this.put_idmeta = {};
      return;
    }
    //if ID Metadata is enabled and the selection is changed by user
    if (this.options && this.metaOnChange) {
      if (this.selectoption1 === "" && this.selectoption2 === "") {
        this.put_idmeta = {};
        return;
      }
      //validate selection
      if (this.selectoption1 === "" || this.selectoption2 === "") {
        this.selecterrorMessage = "You must select or unselect both ID type and Encryption type.";
        this.stopPut = true;
        return;
      }
      //update value based on the selection
      if (this.selectoption1 === "option3") {
        this.idmeta.id_type = 0;
      } else if (this.selectoption1 === "option1") {
        this.idmeta.id_type = 1;
      } else if (this.selectoption1 === "option2") {
        if (this.option4 === "") {
          this.stopPut = true;
          this.selecterrorMessage = "You must select a option."
        }
        if (this.option4 === "IMEI") {
          this.idmeta.id_type = 2;
        } else if (this.option4 === "IDFA") {
          this.idmeta.id_type = 3;
        } else if (this.option4 === "IDFV") {
          this.idmeta.id_type = 4;
        }
      }
      if (this.selectoption2 === "option5") {
        this.idmeta.id_encryption_type = 0;
      } else if (this.selectoption2 === "option6") {
        if (this.option7 === "") {
          this.stopPut = true;
          this.selecterrorMessage = "You must select a option."
        }
        if (this.option7 === "MD5") {
          this.idmeta.id_encryption_type = 1;
        } else if (this.option7 === "SHA25") {
          this.idmeta.id_encryption_type = 2;
        }
      }
      this.put_idmeta = this.idmeta;
    }
  }

  idmeta: Id_Meta = {
    id_encryption_type: Number.MAX_SAFE_INTEGER,
    id_type: Number.MAX_SAFE_INTEGER
  };
  put_idmeta = {};
  isUpdateSubmit: boolean = false;
  isUpdateFailed: boolean = false;
  //putIDmetaUpdate is to update the selection of ID MetaData
  putIDmetaUpdate(data_id: string) {
    this.errorpop = false;
    this.checkMeta();
    if (this.stopPut) {
      this.errorpop = true;
      return;
    }
    this.isUpdateSubmit = true;
    this.dataservice.putMetaUpdate(this.put_idmeta, data_id)
      .subscribe(data => {
        this.reloadCurrentRoute();
      },
        err => {
          this.errorMessage = err.error.message;
          this.isUpdateFailed = true;
        });
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}
