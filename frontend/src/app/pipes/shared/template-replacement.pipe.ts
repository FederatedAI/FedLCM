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

@Pipe({
  name: 'templateReplacement'
})
export class TemplateReplacementPipe implements PipeTransform {

  transform(value: string, ...args: string[] | number[]): unknown {
    if (typeof value !== 'string') {
      throw new Error("The type of the value must be a string");
    }
    for (let i = 0; i < args.length; i++) {
      if (typeof args[i] === 'string' || typeof args[i] === 'number') {
        value = value.replace(/d%/, args[i] + '')
      } else {
        throw new Error("The template replacement value type must be string or numeric");
      }
    }
    return value;
  }

}
