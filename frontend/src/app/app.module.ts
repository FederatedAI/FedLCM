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
import { BrowserModule } from '@angular/platform-browser';
import { HttpClient, HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { AppRoutingModule} from './app-routing.module';
import { AppComponent } from './app.component';
import { ClarityModule } from '@clr/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HeaderComponent } from './components/header/header.component';
import { ContentComponent } from './view/content.component';
import { SideNavComponent } from './components/side-nav/side-nav.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { CertificateMgComponent } from './view/certificate/certificate-mg/certificate-mg.component';
import { InfraComponent } from './view/infra/infra.component';
import { FedMgComponent } from './view/federation/fed-mg/fed-mg.component';
import { EndpointMgComponent } from './view/endpoint/endpoint-mg/endpoint-mg.component';
import { ChartMgComponent } from './view/chart/chart-mg/chart-mg.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FedDetailFateComponent } from './view/federation/fed-detail-fate/fed-detail-fate.component';
import { FedDetailOpneFLComponent } from './view/openfl/fed-detail-openfl/fed-detail-openfl.component';
import { EndpointNewComponent } from './view/endpoint/endpoint-new/endpoint-new.component';
import { ExchangeNewComponent } from './view/federation/exchange-new/exchange-new.component';
import { DirectorNewComponent } from './view/openfl/director-new/director-new.component'
import { DirectorDetailComponent } from './view/openfl/director-detail/director-detail.component'
import { ClusterNewComponent } from './view/federation/cluster-new/cluster-new.component';
import { InfraDetailComponent } from './view/infra/infra-detail/infra-detail.component';
import { EndpointDetailComponent } from './view/endpoint/endpoint-detail/endpoint-detail.component';
import { LoginComponent } from './view/login/login.component';
import { AuthInterceptor } from 'src/utils/auth-interceptor';

import { AlertComponent } from './components/alert/alert.component';
import { SharedModule } from './pipes/shared/shared.module'
import { TranslateModule, TranslateLoader } from '@ngx-translate/core'
import { createTranslateLoader, AppService } from './app.service';
import { ChartDetailComponent } from './view/chart/chart-detail/chart-detail.component';
import { CertificateAuthorityDetailComponent } from './view/certificate/certificate-authority-detail/certificate-authority-detail.component';
import { CertificateDetailComponent } from './view/certificate/certificate-detail/certificate-detail.component';
import { ExchangeDetailComponent } from './view/federation/exchange-detail/exchange-detail.component';
import { ClusterDetailComponent } from './view/federation/cluster-detail/cluster-detail.component'
import { MessageModule } from './components/message/message.module';
import { EventsListComponent } from './components/events-list/events-list.component';
import { TimeOutServiceComponent } from './components/time-out-service/time-out-service.component';
import { FilterComponent } from './components/filter/filter.component';
import { EnvoyDetailComponent } from './view/openfl/envoy-detail/envoy-detail.component';
import { CreateOpenflComponent } from './view/openfl/create-openfl-fed/create-openfl-fed.component';
import { ExchangeClusterUpgradeComponent } from './view/federation/exchange-cluster-upgrade/exchange-cluster-upgrade.component';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    ContentComponent,
    SideNavComponent,
    CertificateMgComponent,
    InfraComponent,
    FedMgComponent,
    EndpointMgComponent,
    ChartMgComponent,
    FedDetailFateComponent,
    FedDetailOpneFLComponent,
    EndpointNewComponent,
    ExchangeNewComponent,
    ClusterNewComponent,
    InfraDetailComponent,
    EndpointDetailComponent,
    LoginComponent,
    AlertComponent,
    ChartDetailComponent,
    CertificateAuthorityDetailComponent,
    ExchangeDetailComponent,
    ClusterDetailComponent,
    EventsListComponent,
    CertificateDetailComponent,
    TimeOutServiceComponent,
    DirectorNewComponent,
    DirectorDetailComponent,
    FilterComponent,
    EnvoyDetailComponent,
    CreateOpenflComponent,
    ExchangeClusterUpgradeComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ClarityModule,
    BrowserAnimationsModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule,
    SharedModule,
    MessageModule,
    TranslateModule.forRoot({// config i8n
      defaultLanguage: 'en',
      loader: {
        provide: TranslateLoader,
        useFactory: createTranslateLoader,
        deps: [HttpClient]
      }
    }),
  ],
  providers: [
    [{ provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true }, AppService]
  ],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
