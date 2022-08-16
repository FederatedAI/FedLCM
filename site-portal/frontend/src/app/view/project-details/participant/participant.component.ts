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
import { ProjectService } from '../../../service/project.service';
import { PARTYSTATUS } from '../../../../config/constant'
import { CustomComparator } from 'src/utils/comparator';

export interface Participant {
  name: string,
  party_id: string,
  desc: string,
  created_time: string,
  status: string
}
export interface PartyUser {
  creation_time: string,
  description: string,
  name: string,
  party_id: number,
  status: number,
  uuid: string,
  selected: boolean
}
export interface ParticipantListResponse {
  code: number,
  data: [],
  message: "success"
}
export interface ProjectDetailResponse {
  code: number,
  data: {},
  message: "success"
}
export interface ProjectDetail {
  auto_approval_enabled: boolean,
  creation_time: string,
  description: string,
  managed_by_this_site: boolean,
  manager: string,
  managing_site_name: string,
  managing_site_party_id: number,
  name: string,
  uuid: string
}

@Component({
  selector: 'app-participant',
  templateUrl: './participant.component.html',
  styleUrls: ['./participant.component.css']
})
export class ParticipantComponent implements OnInit {

  constructor(private route: ActivatedRoute, private projectservice: ProjectService, private router: Router) {
    this.showProjectDetail();
    this.showInvitedParticipantList();
  }
  partyStatus = PARTYSTATUS
  options: boolean = false;
  project: any;
  openModal: boolean = false;
  inviteoption = false;

  ngOnInit(): void {
  }

  routeParams = this.route.parent!.snapshot.paramMap;
  // uuid is project uuid
  uuid = String(this.routeParams.get('id'));
  participantList: any;
  allParticipantList: any;
  allParticipantListResponse: any;
  newAllParticipantList: PartyUser[] = [];
  isShowParticiapantFailed: boolean = false;
  invitedParticipantList: any;
  isPageLoading: boolean = true;
  // showInvitedParticipantList is to get the invited participants list 
  showInvitedParticipantList() {
    this.isPageLoading = true;
    this.projectservice.getParticipantList(this.uuid, false)
      .subscribe((data: ParticipantListResponse) => {
        this.invitedParticipantList = data.data;
        this.isPageLoading = false;
      },
        err => {
          this.isShowParticiapantFailed = true;
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
        }
      );
  }

  // showAllParticipantList is to get all the participants list
  showAllParticipantList() {
    this.openModal = true;
    this.newAllParticipantList = [];
    this.projectservice.getParticipantList(this.uuid, true)
      .subscribe((data: ParticipantListResponse) => {
        this.allParticipantListResponse = data;
        this.allParticipantList = this.allParticipantListResponse.data;
        for (let user of this.allParticipantList) {
          const party: PartyUser =
          {
            creation_time: "",
            description: "",
            name: "",
            party_id: 0,
            status: 0,
            uuid: "",
            selected: false
          };
          party.creation_time = user.creation_time;
          party.description = user.description;
          party.name = user.name;
          party.party_id = user.party_id;
          party.status = user.status;
          party.uuid = user.uuid;
          if (user.status === this.partyStatus.Owner || user.status === this.partyStatus.Pending || user.status === this.partyStatus.Joined) {
            party.selected = true;
          }

          this.newAllParticipantList.push(party);
        }
      },
        err => {
          this.isShowParticiapantFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  isSubmitInvitation: boolean = false;
  isSubmitInvitationFailed: boolean = false;
  noSelected: boolean = true;
  // submitInvitation is to submit the request to invite participant
  submitInvitation() {
    this.isSubmitInvitation = true;
    this.isSubmitInvitationFailed = false;
    for (let party of this.newAllParticipantList) {
      if (party.status == this.partyStatus.Unknown && party.selected) {
        if (this.noSelected) this.noSelected = false;
        this.projectservice.postInvitation(this.uuid, party.description, party.name, party.party_id, party.uuid)
          .subscribe(
            data => {
              this.reloadCurrentRoute();
            },
            err => {
              this.isSubmitInvitationFailed = true;
              this.isSubmitInvitation = false;
              this.errorMessage = err.error.message;
            }
          );
      }
    }
    if (this.noSelected) {
      this.errorMessage = "Please select.";
      this.isSubmitInvitationFailed = true;
    }
  }

  projectDetailResponse: any;
  projectDetail: ProjectDetail = {
    auto_approval_enabled: false,
    creation_time: "",
    description: "",
    managed_by_this_site: false,
    manager: "",
    managing_site_name: "",
    managing_site_party_id: 0,
    name: "",
    uuid: ""
  }
  // showProjectDetail is to get the preject detail to get if the auto approval is enabled
  showProjectDetail() {
    this.projectservice.getProjectDetail(this.uuid)
      .subscribe((data: ProjectDetailResponse) => {
        this.projectDetailResponse = data;
        this.projectDetail = this.projectDetailResponse.data;
        this.options = this.projectDetail.auto_approval_enabled;
      });
  }

  errorMessage: string = "";
  isDeleteSubmitted: boolean = false;
  submitDeleteFailed: boolean = false;
  // deleteParticipant is to request to remove particiapnt (only with the project is managed by current site)
  deleteParticipant(party_uuid: string) {
    this.isDeleteSubmitted= true;
    this.projectservice.deleteParticipant(this.uuid, party_uuid)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.submitDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  cur_party_uuid: string = "";
  isOpenAlertModal: boolean = false;
  // openAlertModal is triggered to open the confirm modal when user want to 'delete Participant'
  openAlertModal(party_uuid: string) {
    this.isOpenAlertModal = true;
    this.cur_party_uuid = party_uuid;
  }
  
  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  // refresh button
  refresh() {
    this.showProjectDetail();
    this.showInvitedParticipantList();
  }

  // comparator for datagrid
  createTimeComparator = new CustomComparator("creation_time", "string");
  roleComparator = new CustomComparator("status", "number");
  statusComparator = new CustomComparator("status", "number");
}
