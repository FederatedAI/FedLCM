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

import { Component, Input, OnInit } from '@angular/core';
import { EventService } from './event.service';
import { EventType, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-events-list',
  templateUrl: './events-list.component.html',
  styleUrls: ['./events-list.component.scss']
})

export class EventsListComponent implements OnInit {
  @Input('entity-uuid') entity_uuid: string = '';
  constructor(private eventService: EventService) { }

  ngOnInit(): void {
    this.showEventlist()
  }

  eventType = EventType;
  constantGather = constantGather

  eventList: any = []
  errorMessage: any;
  isShowEndpointDetailFailed: boolean = false;
  isPageLoading: boolean = true;
  isShowEventlistFailed = false;
  //showEventlist is to get the event list by entity's UUID
  showEventlist() {
    this.isPageLoading = true;
    this.isShowEventlistFailed = false;
    this.eventService.getEventList(this.entity_uuid)
      .subscribe((data: any) => {
        if (data.data) this.eventList = data.data;
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
          this.isShowEndpointDetailFailed = true;
        }
      );
  }

  //refresh is for refresh button
  refresh() {
    this.showEventlist()
  }

}
