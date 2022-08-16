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
import { Router } from '@angular/router';
import { MessageService } from 'src/app/components/message/message.service';

@Component({
  selector: 'app-time-out-service',
  templateUrl: './time-out-service.component.html',
  styleUrls: ['./time-out-service.component.scss']
})
export class TimeOutServiceComponent implements OnInit {

  constructor(private $msg: MessageService, private router: Router) { }

  ngOnInit(): void {
  }

  reload () {
    const redirect = sessionStorage.getItem('lifecycleManager-redirect')
    if (redirect) {
      this.router.navigateByUrl(redirect)
      setTimeout(() => {
        sessionStorage.removeItem('lifecycleManager-redirect')
      }, 1000)
    } else {
      this.router.navigateByUrl('/login')
      this.$msg.warning('serverMessage.default401', 1000)
    }
  }
}
