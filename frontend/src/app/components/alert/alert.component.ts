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

import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'app-alert',
  templateUrl: './alert.component.html',
  styleUrls: ['./alert.component.scss']
})
export class AlertComponent implements OnInit {

  constructor() { }
  @Input()type: 'success' | 'info' | 'warning' | 'danger' = 'info'
  @Input()message: string = ''
  @Input() width:string ='100%'
  @Input() untreatedMargin: string = '0'
  get margin () {
    let margin = ''
    if (this.untreatedMargin.indexOf(',') !== -1) {
      margin = this.untreatedMargin.replace(/,/g, ' ')
    } else {
      margin = this.untreatedMargin
    }
    return margin
  }
  ngOnInit(): void {
  }

}
