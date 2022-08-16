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
import { ClarityModule } from '@clr/angular';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { ProjectDetailsRoutingModule, declarations } from './project-details-routing'
import { CommonModule } from '@angular/common';
import { ShardModule } from 'src/app/shared/shard/shard.module'

@NgModule({
  declarations: [
    ...declarations
  ],
  imports: [
    ProjectDetailsRoutingModule,
    ClarityModule,
    FormsModule,
    CommonModule,
    ShardModule,
    ReactiveFormsModule
  ],
})
export class ProjectDetailModule {}
