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
import { RouterModule, Routes} from '@angular/router';
import { LoginComponent } from './view/login/login.component';
import { CertificateMgComponent } from './view/certificate/certificate-mg/certificate-mg.component';
import { ChartDetailComponent } from './view/chart/chart-detail/chart-detail.component';
import { ChartMgComponent } from './view/chart/chart-mg/chart-mg.component';
import { ClusterNewComponent } from './view/federation/cluster-new/cluster-new.component';
import { ContentComponent } from './view/content.component';
import { EndpointDetailComponent } from './view/endpoint/endpoint-detail/endpoint-detail.component';
import { EndpointMgComponent } from './view/endpoint/endpoint-mg/endpoint-mg.component';
import { EndpointNewComponent } from './view/endpoint/endpoint-new/endpoint-new.component';
import { ExchangeNewComponent } from './view/federation/exchange-new/exchange-new.component';
import { FedDetailFateComponent } from './view/federation/fed-detail-fate/fed-detail-fate.component';
import { FedDetailOpneFLComponent } from './view/openfl/fed-detail-openfl/fed-detail-openfl.component'
import { FedMgComponent } from './view/federation/fed-mg/fed-mg.component';
import { InfraDetailComponent } from './view/infra/infra-detail/infra-detail.component';
import { InfraComponent } from './view/infra/infra.component';
import { CertificateAuthorityDetailComponent } from './view/certificate/certificate-authority-detail/certificate-authority-detail.component'
import { CertificateDetailComponent } from './view/certificate/certificate-detail/certificate-detail.component'
import { ExchangeDetailComponent } from './view/federation/exchange-detail/exchange-detail.component'
import { ClusterDetailComponent } from './view/federation/cluster-detail/cluster-detail.component'
import { TimeOutServiceComponent } from './components/time-out-service/time-out-service.component'
import { DirectorNewComponent } from './view/openfl/director-new/director-new.component'
import { DirectorDetailComponent } from './view/openfl/director-detail/director-detail.component'
import { EnvoyDetailComponent } from './view/openfl/envoy-detail/envoy-detail.component';
import { ExchangeClusterUpgradeComponent } from './view/federation/exchange-cluster-upgrade/exchange-cluster-upgrade.component';
import { AuthService } from './services/common/auth.service';
import { RouterGuard } from './router-guard';

const routes: Routes = [
  {
    path: 'login',
    component: LoginComponent
  },
  {
    path: '',
    component: ContentComponent,
    data: {
      preload: true
    },
    children: [
      {
        path: '',
        redirectTo: 'federation',
        pathMatch: 'full'
      },
      {
        path: 'federation',
        data: {
          preload: true,
        },
        component: FedMgComponent
      },
      {
        path: 'infra',
        data: {
          preload: true
        },
        component: InfraComponent
      },
      {
        path: 'endpoint',
        data: {
          preload: true
        },
        component: EndpointMgComponent
      },
      {
        path: 'chart',
        data: {
          preload: true
        },
        component: ChartMgComponent
      },
      {
        path: 'certificate',
        data: {
          preload: true
        },
        component: CertificateMgComponent
      },
      {
        path: 'certificate/authority/:id',
        data: {
          preload: true
        },
        component: CertificateAuthorityDetailComponent
      },
      {
        path: 'certificate/detail/:id',
        data: {
          preload: true
        },
        component: CertificateDetailComponent
      },

      {
        path: 'federation/fate/:id',
        data: {
          preload: true
        },
        component: FedDetailFateComponent,
      },
      {
        path: 'federation/openfl/:id',
        data: {
          preload: true
        },
        component: FedDetailOpneFLComponent,
        canActivate: [RouterGuard],
      },
      {
        path: 'endpoint/new',
        data: {
          preload: true
        },
        component: EndpointNewComponent
      },
      {
        path: 'federation/fate/:id/exchange/new',
        data: {
          preload: true
        },
        component: ExchangeNewComponent
      },
      {
        path: 'federation/openfl/:id/director/new',
        data: {
          preload: true
        },
        component: DirectorNewComponent,
        canActivate: [RouterGuard],
      },
      {
        path: 'federation/fate/:id/cluster/new',
        data: {
          preload: true
        },
        component: ClusterNewComponent
      },
      {
        path: 'federation/fate/:id/exchange/detail/:exchange_uuid',
        data: {
          preload: true
        },
        component: ExchangeDetailComponent
      },
      {
        path: 'federation/openfl/:id/director/detail/:director_uuid',
        data: {
          preload: true
        },
        component: DirectorDetailComponent,
        canActivate: [RouterGuard],
      },
      {
        path: 'federation/fate/:id/cluster/detail/:cluster_uuid',
        data: {
          preload: true
        },

        component: ClusterDetailComponent
      },
      {
        path: 'federation/fate/:id/detail/:uuid/:version/:name/upgrade',
        data: {
          preload: true
        },

        component: ExchangeClusterUpgradeComponent
      },
      {
        path: 'federation/openfl/:id/envoy/detail/:envoy_uuid',
        data: {
          preload: true
        },

        component: EnvoyDetailComponent,
        canActivate: [RouterGuard],
      },
      {
        path: 'infra-detail/:id',
        data: {
          preload: true
        },
        component: InfraDetailComponent
      },
      {
        path: 'endpoint-detail/:id',
        data: {
          preload: true
        },
        component: EndpointDetailComponent
      },
      {
        path: 'chart-detail/:id',
        data: {
          preload: true
        },
        component: ChartDetailComponent
      },
      {
        path: 'timeout',
        data: {
          preload: true
        },
        component: TimeOutServiceComponent
      }
    ]
  },
];
@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
  providers:[RouterGuard, AuthService]
})

export class AppRoutingModule { 
}

