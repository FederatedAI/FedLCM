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
import { RouterModule, Routes } from '@angular/router';
import { ProjectDetailsComponent } from './project-details.component';
import { InfoComponent } from './info/info.component'
import { DataComponent } from './data/data.component';
import { JobComponent } from './job/job.component';
import { ModelComponent } from './model/model.component';
import { ParticipantComponent } from './participant/participant.component';


const routes: Routes = [
  {
    path:'',
    component: ProjectDetailsComponent,
    children: [
      { path: '', redirectTo: 'info', pathMatch: 'full' },
      { path:'info', component: InfoComponent},
      { path:'job', component: JobComponent},
      { path:'data', component: DataComponent},
      { path:'model', component: ModelComponent},
      { path:'participant', component: ParticipantComponent}
      ]
  }
]
export const declarations = [
  ProjectDetailsComponent,
  InfoComponent,
  JobComponent,
  DataComponent,
  ModelComponent,
  ParticipantComponent
]
@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ProjectDetailsRoutingModule { }