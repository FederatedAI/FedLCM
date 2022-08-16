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

import { Component, OnInit, ViewChild } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { FedService } from 'src/app/services/federation-fate/fed.service';
import { OpenflService } from 'src/app/services/openfl/openfl.service'
import { ValidatorGroup } from 'src/utils/validators'
import { CreateOpenflComponent } from 'src/app/view/openfl/create-openfl-fed/create-openfl-fed.component'
import { AuthService } from 'src/app/services/common/auth.service';
@Component({
  selector: 'app-fed-mg',
  templateUrl: './fed-mg.component.html',
  styleUrls: ['./fed-mg.component.scss']
})
export class FedMgComponent implements OnInit {
  @ViewChild('create_openfl') create_openfl!: CreateOpenflComponent
  federationList: any[] = [];
  openModal: boolean = false;
  disabled: boolean = true;
  //default federation type is 'fate'
  federationType = 'fate'
  //fedInformationForm is form to create a new federation
  fedInformationForm = this.fb.group(
    ValidatorGroup([
      {
        name: 'fedname',
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
        name: 'domain',
        value: '',
        type: ['fqdn']
      }
    ])
  )

  constructor(private fb: FormBuilder, private fedservice: FedService, private router: Router, private openflService: OpenflService, private authService: AuthService) {
    this.getLCMServiceStatus();
    this.showFedList();
  }

  ngOnInit(): void {
  }
  fate: boolean = true;
  openfl: boolean = false;
  suffix: string = "";
  //setRadioDisplay is trigger when the selection of federation type is changed
  setRadioDisplay(val: any) {
    if (val == "fate") {
      this.fate = true;
      this.openfl = false;
    }
    if (val == "openfl") {
      this.fate = false;
      this.openfl = true;
    }
  }
  //onOpenModal is to open the 'Create a new federation' modal
  onOpenModal() {
    this.federationType = "fate"
    this.openModal = true;
  }
  //resetModal is for resetting the modal when close
  resetModal() {
    this.fedInformationForm.reset();
    this.openModal = false;
  }

  errorMessage = "Service Error!"
  isShowFedFailed: boolean = false;
  Fedlist: any;
  isPageLoading: boolean = true;
  //showFedList is to show federation list
  showFedList() {
    this.isShowFedFailed = false;
    this.isPageLoading = true;
    this.fedservice.getFedList()
      .subscribe((data: any) => {
        this.federationList = data.data;
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isShowFedFailed = true
          this.isPageLoading = false
        });
  }

  isCreatedSubmit = false;
  isCreatedFailed = false;
  //submitDisable returns if disabled the submit button of create a new federation
  get submitDisable() {
    if (this.federationType === 'fate') {
      if (this.fedInformationForm.valid) {
        return false
      } else {
        return true
      }
    } else {
      if (!this.create_openfl?.openflForm.get('customize')?.value) {
        return !this.create_openfl?.customizeFalseValidate()
      } else {
        return !this.create_openfl?.customizeTrueValidate()
      }
    }
  }
  
  //createNewFed is to submit the 'create federation' request
  createNewFed() {
    this.isCreatedFailed = false;
    this.isCreatedSubmit = true;
    // fate
    if (this.federationType === 'fate') {
      if (this.fedInformationForm.valid) {
        const fedInfo = {
          description: this.fedInformationForm.get('description')?.value,
          domain: this.fedInformationForm.get('domain')?.value,
          name: this.fedInformationForm.get('fedname')?.value
        }
        this.fedservice.createFed(fedInfo)
          .subscribe(
            data => {
              this.isCreatedFailed = false;
              this.reloadCurrentRoute();
            },
            err => {
              this.errorMessage = err.error.message;
              this.isCreatedFailed = true;
            }
          );
      } else {
        this.errorMessage = "Invalid input";
        this.isCreatedFailed = true;
      }
    } else {// openfl
      this.create_openfl.createNewOpenfl().subscribe(
        data => {
          this.isCreatedFailed = false;
          this.reloadCurrentRoute();
        },
        err => {
          this.errorMessage = "Invalid input";
          this.isCreatedFailed = true;
        }
      )
    }

  }

  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataList: string = '';
  //openDeleteConfrimModal is to initial variables when open the modal of 'Delete federation'
  openDeleteConfrimModal(uuid: string, fedType: string) {
    this.pendingFed = uuid;
    this.pendingFedType = fedType
    this.isDeleteFailed = false;
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
  }

  pendingFed: string = "";
  pendingFedType = ''
  //deleteFed is to delete federation
  deleteFed() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    if (this.pendingFedType === 'FATE') {
      this.fedservice.deleteFed(this.pendingFed)
        .subscribe(() => {
          this.reloadCurrentRoute();
        },
          err => {
            this.isDeleteFailed = true;
            this.errorMessage = err.error.message;
          });
    } else {
      this.openflService.deleteOpenflFederation(this.pendingFed)
        .subscribe(() => {
          this.reloadCurrentRoute();
        },
          err => {
            this.isDeleteFailed = true;
            this.errorMessage = err.error.message;
          });
    }
  }

  experimentEnabled = false;
  isGetLCMStatusFailed = false;
  //getLCMServiceStatus is to get the experiment is enabled or not
  getLCMServiceStatus() {
    this.isShowFedFailed = false;
    this.isPageLoading = true;
    this.authService.getLCMServiceStatus()
      .subscribe((data: any) => {
        this.experimentEnabled = data?.experiment_enabled;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isGetLCMStatusFailed = true
        });
  }

  //refresh is for refresh button
  refresh() {
    this.showFedList();
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/infra', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}
