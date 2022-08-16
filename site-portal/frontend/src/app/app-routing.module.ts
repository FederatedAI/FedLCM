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
import { DataDetailsComponent } from './view/data-details/data-details.component';
import { DataMgComponent } from './view/data-mg/data-mg.component';
import { ModelDetailComponent } from './view/model-detail/model-detail.component';
import { ModelMgComponent } from './view/model-mg/model-mg.component';
import { ProjectMgComponent } from './view/project-mg/project-mg.component';
import { SiteConfigComponent } from './view/site-config/site-config.component';
import { LogInComponent } from './view/log-in/log-in.component';
import { UserMgComponent } from './view/user-mg/user-mg.component';
import { HomeComponent } from './view/home/home.component'
import { SelectivePreloadingStrategy } from 'src/utils/selective-preloading-strategy';
const routes: Routes = [
  {
    path: 'login',
    component: LogInComponent
  },
  {
    path: '',
    component: HomeComponent,
    data: {
      preload: true
    },
    children: [
      {
        path: '',
        redirectTo: 'project-management',
        pathMatch: 'full'
      },
      {
        path: 'model-management',
        data: {
          preload: true
        },
        component: ModelMgComponent
      },
      {
        path: 'site-configuration',
        data: {
          preload: true
        },
        component: SiteConfigComponent,
      },
      {
        path: 'user-management',
        data: {
          preload: true
        },
        component: UserMgComponent
      },
      {
        path: 'site-configuration',
        data: {
          preload: true
        },
        component: SiteConfigComponent
      },
      {
        path: 'model-detail/:id',
        data: {
          preload: true
        },
        component: ModelDetailComponent
      },
      {
        path: 'data-management',
        data: {
          preload: true
        },
        component: DataMgComponent
      },
      {
        path: 'data-detail/:data_id',
        data: {
          preload: true
        },
        component: DataDetailsComponent
      },
      {
        path: 'project-management',
        data: {
          preload: true
        },
        component: ProjectMgComponent
      },
      {
        path: 'project-management/project-detail/:id',
        loadChildren: () => import('./view/project-details/project-detail.module').then(m => m.ProjectDetailModule)
      },
      {
        path: 'project-management/project-detail/:projid/job/job-detail/:jobid',
        loadChildren: () => import('./view/job-detail/job-detail.module').then(m => m.JobDetailModule)
      },
      {
        path: 'project-management/project-detail/:id/job/new-job',
        loadChildren: () => import('./view/job-new/job-new.module').then(m => m.JobNewModule)
      }
    ]
  },

];

@NgModule({
  imports: [RouterModule.forRoot(routes, {
    preloadingStrategy: SelectivePreloadingStrategy
  })],
  exports: [RouterModule]
})
export class AppRoutingModule { }
