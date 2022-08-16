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
import { FormBuilder } from '@angular/forms';
import { CertificateMgService } from 'src/app/services/common/certificate.service'
import { CertificateType } from '../certificate-model';
import { constantGather } from 'src/utils/constant'
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-certificate-mg',
  templateUrl: './certificate-mg.component.html',
  styleUrls: ['./certificate-mg.component.scss']
})
export class CertificateMgComponent implements OnInit {
  constantGather = constantGather
  certificatelist: CertificateType[] = [];
  uuid = ''
  ca: any
  openModal: boolean = false;
  selected: any = [];

  type: string = '';
  institution: string = '';
  authority: string = '';
  note: string = '';
  startdate: number = Date.now();;
  enddate: number = 0;
  today: number = Date.now();
  newCertificateForm = this.fb.group({
    type: [''],
    name: [''],
    authority: [''],
    note: [''],
    startdate: [''],
    enddate: [''],
    seriesNumber: [''],
    server: ['']
  });

  openDeleteModal = false
  errorMessage = "Service Error!"
  isPageLoading = true;

  constructor(private fb: FormBuilder, private certificateService: CertificateMgService, private router: Router, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.getCAinfo();
    this.getCertificateList();

  }
  onOpenModal() {
    this.openModal = true;
  }

  //Currently, 'Add a new certificate' is not supported
  // resetModal() {
  //   this.newCertificateForm.reset();
  //   this.openModal = false;
  // }

  //getCAinfo is to get certificate authority information
  getCAFailed = false
  getCAinfo() {
    this.isPageLoading = true
    this.getCAFailed = false
    this.certificateService.getCertificateAuthority().subscribe(
      data => {
        if (data.data) this.ca = data.data
        this.isPageLoading = false;
      },
      err => {
        this.isPageLoading = false
        this.getCAFailed  = true
        if (err.error.message) this.errorMessage = err.error.message
      }
    )
  }

  //getCertificateList is to get certificate list
  getCertificateListFailed = false
  getCertificateList() {
    this.getCertificateListFailed = false
    this.certificateService.getCertificateList().subscribe(
      data => {
        if (data.data) {
          this.certificatelist = data.data.map(el => {
            el.select = false
            return el
          })
        } else {
          this.certificatelist = []
        }
      },
      err => {
        this.isPageLoading = false;
        this.getCertificateListFailed = true
        if (err.error.message) this.errorMessage = err.error.message
      }
    )
  }

  nameSortFlag = false
  nameSort(type: string) {
    this.nameSortFlag = !this.nameSortFlag
    if (this.nameSortFlag) {
      this.certificatelist.sort(reverse(type))
    } else {
      this.certificatelist.sort(order(type))
    }
  }

  //get allSelect is to get the selected list
  get allSelect() {
    if (this.certificatelist.length > 0) {
      return this.certificatelist.every(el => el.select === true)
    } else {
      return false
    }
  }
  //set allSelect is to select all certificate
  set allSelect(val) {
    this.certificatelist.forEach(el => this.checkCertificateBindings(el) ? el.select = val : el.select = false)
  }

  //disableAllSelect is the flag to disable "select all" checkbox
  get disableAllSelect() {
    if (this.certificatelist.length > 0) return this.certificatelist.every(el => el.bindings.length > 0)
    return true
  }

  //openDeleteConfrimModal is to initial variables when open the modal of 'Delete Certificate'
  openDeleteConfrimModal() {
    this.openDeleteModal = true
    this.isDeleteCertificateAllSuccess = false
  }

  //deleteSelectedCertificat is to delete the selected certificate
  async deleteSelectedCertificate() {
    if (this.selectedCertificateList?.length > 0) {
      await this.submitDeleteCertificate()
      if (this.deleteCertificateAllSuccess()) this.reloadCurrentRoute()
    }
  }

  // submitDeleteCertificate is to submit http request for deletion with synchronization
  async submitDeleteCertificate() {
    for (let certificate of this.selectedCertificateList) {
      certificate.deleteFailed = false
      certificate.select = !certificate.deleteSuccess
      certificate.deleteSubmit = true
      if (!certificate.deleteSuccess && certificate.select) {
        await this.certificateService.deleteCertificate(certificate.uuid)
          .toPromise().then(() => {
            certificate.deleteFailed = false;
            certificate.deleteSuccess = true;
          },
          err => {
            this.openDeleteModal = true
            certificate.deleteFailed = true;
            certificate.deleteSuccess = false;
            certificate.errorMessage = err.error.message
          });
      }
    }
  }

  //checkCertificateBindings is to check if the certificate contains binding participants
  checkCertificateBindings(certificate: CertificateType): boolean {
    if (certificate) {
      return !(certificate?.bindings?.length > 0)
    }
    return false
  }

  get selectedCertificateList() {
    const selectedList: CertificateType[] = []
    this.certificatelist.forEach(el => {
      el.select ? selectedList.push(el) : null
    })
    return selectedList
  }

  get deleteBtnDisabled() {
    return !(this.selectedCertificateList?.length > 0)
  }

  isDeleteCertificateAllSuccess = false;
  //deleteCertificateAllSuccess is to return if multiple deletion are all successful
  deleteCertificateAllSuccess() {
    for (const certificate of this.selectedCertificateList) {
      if (!certificate.deleteSuccess) {
        this.isDeleteCertificateAllSuccess = false
        return this.isDeleteCertificateAllSuccess
      }
    }
    this.isDeleteCertificateAllSuccess = true
    return this.isDeleteCertificateAllSuccess
  }

  //toDetail is redirect to certificate or CA detail page
  toDetail(uuid: string, type?: boolean) {
    if (type) {
      this.router.navigate(['/certificate/detail/', uuid]);
    } else {
      this.router.navigate(['/certificate/authority', uuid]);
    }
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}
// positive order function
function order(propertyName: string) {
  return function (obj1: any, obj2: any) {
    var value1 = obj1[propertyName];
    var value2 = obj2[propertyName];
    if (value1 < value2) {
      return -1;
    } else if (value1 > value2) {
      return 1;
    } else {
      return 0;
    }
  }
}
// positive reverse function
function reverse(propertyName: string) {
  return function (obj1: any, obj2: any) {
    var value1 = obj1[propertyName];
    var value2 = obj2[propertyName];
    if (value1 < value2) {
      return 1;
    } else if (value1 > value2) {
      return -1;
    } else {
      return 0;
    }
  }
}