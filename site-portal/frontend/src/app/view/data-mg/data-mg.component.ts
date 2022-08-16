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

import { Component, OnInit, ChangeDetectorRef, OnDestroy, ViewChild } from '@angular/core';
import { UploadFileComponent } from '../../components/upload-file/upload-file.component'
import { FormBuilder, FormGroup } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import '@cds/core/icon/register.js';
import { addTextIcon, ClarityIcons, searchIcon, uploadIcon } from '@cds/core/icon';
import '@cds/core/file/register.js';
import { DataService } from '../../service/data.service';
import { MessageService } from '../../components/message/message.service'
import * as fileSaver from 'file-saver';
import { ValidatorGroup } from '../../../config/validators'
import { CustomComparator } from 'src/utils/comparator';
import { SiteService } from 'src/app/service/site.service';
import { HttpEvent, HttpEventType, HttpHeaderResponse, HttpResponse } from '@angular/common/http';

ClarityIcons.addIcons(searchIcon, addTextIcon, uploadIcon);

export interface DataListResponse {
  code: number;
  message: string;
  data: DataElement[];
}
export interface DataColumnResponse {
  code: number;
  message: string;
  data: String[];
}
export interface DataElement {
  creation_time: string;
  data_id: string;
  feature_size: number;
  name: string;
  sample_size: number;
}

@Component({
  selector: 'app-data-mg',
  templateUrl: './data-mg.component.html',
  styleUrls: ['./data-mg.component.css']
})


export class DataMgComponent implements OnInit, OnDestroy {

  form: FormGroup;
  @ViewChild('file') fileCof !: UploadFileComponent
  constructor(private fb: FormBuilder, private dataservice: DataService, private route: ActivatedRoute,
    private router: Router, private changeDetectorRef: ChangeDetectorRef, private msg: MessageService, private siteService: SiteService) {
    this.showDataList();
    this.form = this.fb.group(
      ValidatorGroup([
        {
          name: 'name',
          value: '',
          type: ['word'],
          max: 20,
          min: 2
        },
        {
          name: 'description',
          type: ['']
        },
        {
          name: 'file',
          type: ['']
        },
      ])
    );
  }

  ngOnInit(): void {

  }
  ngOnDestroy(): void {
    this.msg.close()
  }

  timeComparator = new CustomComparator("creation_time", "string");
  datalistresponse: any;
  public datalist: any[] = []
  // Save all data
  public storageDataList: any[] = []
  isPageLoading: boolean = true;
  isShowDataFailed: boolean = false;
  //showDataList is to get the current data list
  showDataList() {
    this.isPageLoading = true;
    this.dataservice.getDataList()
      .subscribe((data: DataListResponse) => {
        this.datalistresponse = data;
        //default descending order
        this.datalist = this.datalistresponse.data?.sort((n1: DataElement, n2: DataElement) => { return this.timeComparator.compare(n1, n2) });
        this.storageDataList = this.datalist;
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowDataFailed = true;
          this.isPageLoading = false;
          this.isPageLoading = false;
        });
  }

  //filterDataList is triggered when user set up the filter condition
  filterDataList(data: any) {
    this.datalist = []
    if (data.searchValue.trim() === '' && data.eligibleList.length < 1) {
      this.datalist = this.storageDataList
    } else {
      data.eligibleList.forEach((el: any) => {
        this.datalist.push(el)
      })
    }
  }
  //showFilter is to display the filtered data list
  isShowfilter: boolean = false;
  showFilter() {
    this.isShowfilter = !this.isShowfilter;
  }
  //refresh button
  refresh() {
    this.showDataList();
    this.isShowfilter = false;
  }

  name: string = "";
  description: string = "";
  uploadfile: string = "";
  openModal = false;
  filterSearchValue = "";
  submitted: boolean = false;
  loading: boolean = false;
  isUploadFailed: boolean = false;
  errorMessage: string = "";
  progress = 0
  //uploadData is to submit upload data request
  uploadData() {
    this.submitted = true;
    this.loading = true;
    this.form.get('file')?.setValue(this.fileCof.file)
    //validate the form and input
    if (this.form.get('name')!.value === '') {
      this.errorMessage = "Name can not be empty.";
      this.isUploadFailed = true;
      return;
    }
    if (!this.fileCof.isUploaded) {
      this.errorMessage = "Please upload a file.";
      this.isUploadFailed = true;
      return;
    }
    if (!this.form.valid) {
      this.errorMessage = 'Invalid information.';
      this.isUploadFailed = true;
      return;
    }
    var formData: any = new FormData();
    formData.append('name', this.form.get('name')!.value);
    formData.append('description', this.form.get('description')!.value);
    formData.append('file', this.form.get('file')!.value);
    this.dataservice.uploadData(formData).subscribe(
      (data: HttpEvent<FileType> | HttpHeaderResponse | HttpResponse<any>) => {
        if (data.type === HttpEventType.UploadProgress) {
          if (data.total && data.total > 0) {
            this.progress = data.loaded / data.total * 100
          }
        }
        if (data.type === HttpEventType.Response) {
          if (data.status === 200) {
            this.msg.success('serverMessage.upload200', 1000)
            this.isUploadFailed = false;
            this.loading = false;
            setTimeout(() => {
              this.reloadCurrentRoute();
            }, 500)
          } else {
            this.progress = 0
          }
        }
      },
      err => {
        this.errorMessage = err.error.message;
        this.isUploadFailed = true;


      }
    );
  }
  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataId: string = '';
  //openConfirmModal is triggered to open the modal for user to confirm the deletion
  openConfirmModal(data_id: string) {
    this.pendingDataId = data_id;
    this.openDeleteModal = true;
  }
  //deleteData is to submit the request to delete tha data
  deleteData() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.dataservice.deleteData(this.pendingDataId)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  isDownloadSubmit: boolean = false;
  isDownloadFailed: boolean = false;
  //downloadData is to down load data
  downloadData(data_id: string, data_name: string) {
    this.isDownloadSubmit = true;
    this.isDownloadFailed = false;
    this.dataservice.downloadDataDetail(data_id)
      .subscribe((data: any) => {
        let blob: any = new Blob([data], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        fileSaver.saveAs(blob, data_name + '.csv');
      }), (error: any) => {
        this.isDownloadFailed = true;
        this.errorMessage = 'Error downloading the file';
      }
      , () => console.info('Data downloaded successfully');
  };

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  //openUploadDataModal is to open 'update local data modal'
  openUploadDataModal() {
    this.openModal = true
    //send request to confirm the current authenticate is not expired
    this.siteService.getCurrentUser().subscribe(
      data => { },
      err => { }
    )
  }
}
interface FileType {
  type: number
  total?: number
  loaded?: number
}



