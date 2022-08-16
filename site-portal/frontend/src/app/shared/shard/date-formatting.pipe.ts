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

import { Pipe, PipeTransform } from '@angular/core';
import * as moment from 'moment'
@Pipe({
  name: 'dateFormatting',
  pure: false
})
export class DateFormattingPipe implements PipeTransform {

  transform(value: any, ...args: unknown[]): unknown {
    let res = ''
    const lang = localStorage.getItem('Site-Portal-Language') || 'en'
    if (lang === 'zh_CN') {
      res = moment(value).format('YYYY年MM月DD日 HH时mm分ss秒')
    } else {
      res = moment(value).format('ll, LTS')
    }
    return res;
  }

}
