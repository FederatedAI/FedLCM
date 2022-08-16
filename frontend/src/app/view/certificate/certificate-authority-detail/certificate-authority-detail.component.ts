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
import { ValidatorGroup } from 'src/utils/validators'
import { ActivatedRoute, Router } from '@angular/router'
import { CertificateMgService } from 'src/app/services/common/certificate.service'
import { constantGather } from 'src/utils/constant'

@Component({
  selector: 'app-certificate-detail',
  templateUrl: './certificate-authority-detail.component.html',
  styleUrls: ['./certificate-authority-detail.component.scss']
})
export class CertificateAuthorityDetailComponent implements OnInit {
  //form is for Certificate Authority configuration information
  form = this.fb.group(
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
        name: 'type',
        type: [''],
        value: 1
      },
      {
        name: 'embedding',
        value: '',
        type: [''],
      },
      {
        name: 'url',
        value: '',
        type: ['internet']
      },
      {
        name: 'provisionerName',
        value: '',
        type: ['']
      },
      {
        name: 'provisionerPassword',
        value: '',
        type: ['']
      },
      {
        name: 'pem',
        type: ['']
      }
    ])

  )
  isAdd = false
  isUpdate = false
  uuid = ''
  isShowDetailFailed = false
  errorMessage = ''
  isUpdateBtn = true
  submiting = false
  constructor(private fb: FormBuilder, private route: ActivatedRoute, private certificateService: CertificateMgService, private router: Router) { }

  ngOnInit(): void {
    this.route.params.subscribe(p => {
      if (p.id && p.id === 'new') {
        this.isAdd = true
        //enable form to edit
        this.setDisabled(false)
      } else {
        this.uuid = p.id
        this.isAdd = false
        //disable form to edit
        this.setDisabled(true)
        this.getCAConfig()
      }
    })
  }
  constantGather = constantGather

  isGetEmbeddingCAConfigFailed = false;
  //getEmbeddingCAConfig is for get the built-in CA config info if lifecycle manager is deployed with step-ca service
  getEmbeddingCAConfig() {
    this.isGetEmbeddingCAConfigFailed = false;
    this.certificateService.getEmbeddingCAConfig().subscribe(
      data => {
        if (data.data) {
          this.form.controls["provisionerName"].setValue(data.data.provisioner_name)
          this.form.controls["provisionerPassword"].setValue(data.data.provisioner_password)
          this.form.controls["pem"].setValue(data.data.service_cert_pem)
          this.form.controls["url"].setValue("https://" + data.data.service_url + ":9000")
        }
      },
      err => {
        this.isGetEmbeddingCAConfigFailed = true
        this.errorMessage = err.error.message
      }
    )
  }

  //onEmbeddingChange is to reset the CA config form to when selection is changed
  onEmbeddingChange() {
    this.isGetEmbeddingCAConfigFailed = false;
    if (this.form.get("embedding")?.value === "embedding") {
      this.getEmbeddingCAConfig()
    } else {
      if (this.isAdd) {
        this.form.controls["provisionerName"].setValue('')
        this.form.controls["provisionerPassword"].setValue('')
        this.form.controls["pem"].setValue('')
        this.form.controls["url"].setValue('')
      } else {
        this.form.controls["provisionerName"].setValue(this.caDetail?.provisionerName)
        this.form.controls["provisionerPassword"].setValue(this.caDetail?.provisionerPassword)
        this.form.controls["pem"].setValue(this.caDetail?.pem)
        this.form.controls["url"].setValue(this.caDetail?.url)
      }
    }
  }
  get useStepCAType() {
    return this.form.get('type')?.value === "1"
  }

  get updateFlag() {
    return this.isUpdate
  }

  set updateFlag(val) {
    if (val) {
      //enable form to update
      this.setDisabled(false)
    }
    this.isUpdate = val
  }
  get submit() {
    if (this.isAdd) {
      return this.form.valid
    } else {
      if (this.updateFlag) {
        return this.form.valid
      } else {
        return false
      }
    }
  }
  //setDisabled is to set CA config form disabled or not
  setDisabled(opt: boolean) {
    for (const key in this.form.controls) {
      const val = this.form.get(key)
      if (val) {
        if (opt) {
          val.disable()
        } else {
          val.enable()
        }
      }
    }
  }
  caDetail: any
  caStatus = 0
  caStatusMessage = ""
  getCaDetailFailed = false;
  isPageLoading = false;
  //getCAConfig is to getting the CA config info and fill the form with
  getCAConfig() {
    this.isPageLoading = true;
    this.getCaDetailFailed = false;
    this.certificateService.getCertificateAuthority().subscribe(
      data => {
        if (data.data) {
          const value = {
            provisionerName: data.data.config.provisioner_name,
            provisionerPassword: data.data.config.provisioner_password,
            pem: data.data.config.service_cert_pem,
            url: data.data.config.service_url,
            description: data.data.description,
            name: data.data.name,
            type: data.data.type + '',
            embedding: ""
          }
          this.form.setValue(value)
          this.caDetail = value
          this.caStatus = data.data.status
          this.caStatusMessage = data.data.status_message
          this.isPageLoading = false;
        }
      },
      err => {
        this.getCaDetailFailed = true;
        this.isPageLoading = false;
      }
    )
  }
  //submitCAConfig is to create/update the CA configuration
  submitCAConfig(value: any) {
    this.submiting = true
    const data = {
      config: {
        provisioner_name: value.provisionerName?.trim(),
        provisioner_password: value.provisionerPassword?.trim(),
        service_cert_pem: value.pem?.trim(),
        service_url: value.url?.trim()
      },
      description: value.description,
      name: value.name,
      type: value.type * 1
    }
    if (this.isAdd) {
      this.certificateService.createCertificateAuthority(data).subscribe(
        data => {
          this.isUpdateBtn = true
          this.updateFlag = false
          this.isAdd = false
          this.setDisabled(true)
          this.submiting = false
          this.isShowDetailFailed = false
          this.router.navigateByUrl('/certificate')
        },
        err => {
          this.isShowDetailFailed = true
          this.errorMessage = err.error.message
          this.submiting = false
        }
      )
    } else {
      this.certificateService.updateCertificateAuthority(this.uuid, data).subscribe(
        data => {
          this.isUpdateBtn = true
          this.updateFlag = false
          this.isAdd = false
          this.setDisabled(true)
          this.submiting = false
          this.isShowDetailFailed = false
          this.reloadCurrentRoute()
        },
        err => {
          this.isShowDetailFailed = true
          this.errorMessage = err.error.message
          this.submiting = false
        }
      )
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
interface AbstractControl { onlySelf?: boolean | undefined; emitEvent?: boolean | undefined; }
