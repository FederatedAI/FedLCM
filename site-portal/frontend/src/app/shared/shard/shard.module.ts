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

import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TemplateReplacementPipe } from './template-replacement.pipe'
import { TranslateModule } from '@ngx-translate/core';
import { DateFormattingPipe } from './date-formatting.pipe';
import { HighJsonComponent } from '../../components/high-json/high-json.component'


@NgModule({
  declarations: [
    TemplateReplacementPipe,
    DateFormattingPipe,
    HighJsonComponent
  ],
  imports: [
    CommonModule,
    TranslateModule
  ],
  exports: [TemplateReplacementPipe, TranslateModule, DateFormattingPipe, HighJsonComponent]
})
export class ShardModule { }
