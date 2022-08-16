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
import { CertificateMgService } from 'src/app/services/common/certificate.service'
import { CertificateType } from '../certificate-model';
import { ActivatedRoute } from '@angular/router'
import { constantGather, CerificateServiceType } from 'src/utils/constant';
@Component({
  selector: 'app-certificate-detail',
  templateUrl: './certificate-detail.component.html',
  styleUrls: ['./certificate-detail.component.scss']
})
export class CertificateDetailComponent implements OnInit {

  constructor(private certificateService: CertificateMgService, private route: ActivatedRoute) { }
  isPageLoading = true
  isShowDetailFailed = false
  errorMessage = "Service Error!"
  cerificateType = CerificateServiceType
  constantGather = constantGather
  certificateDetail: CertificateType = {
    name: '',
    serial_number: '',
    expiration_date: '',
    common_name: '',
    uuid: '',
    bindings: [],
    deleteFailed: false,
    deleteSuccess: false,
    deleteSubmit: false,
    errorMessage: ""
  }
  private uuid: string = ''
  ngOnInit(): void {
    this.route.params.subscribe(value => {
      this.uuid = value.id
      this.getCertificateDetail()
    })
  }
  //getCertificateDetailis to get the certificate detail info
  getCertificateDetail() {
    this.isPageLoading = true;
    //firstly get the certificate list due to there is no 'get certificate datail' API
    this.certificateService.getCertificateList().subscribe(
      (data: { data: any[]; }) => {
        if (data.data) {
          this.isPageLoading = false;
          //find the spcific certificate information by certificate UUID
          this.certificateDetail = data.data.find((el: any) => el.uuid === this.uuid)
          if (!this.certificateDetail) this.errorMessage = "record not found"
        }
      }
    )
  }
}
