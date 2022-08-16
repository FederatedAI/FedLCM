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

import { NgModule, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ClarityModule } from '@clr/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HeaderComponent } from './components/header/header.component';
import { SidenavComponent } from './components/sidenav/sidenav.component';
import { ModelMgComponent } from './view/model-mg/model-mg.component';
import { SiteConfigComponent } from './view/site-config/site-config.component';
import { UserMgComponent } from './view/user-mg/user-mg.component';
import { ModelDetailComponent } from './view/model-detail/model-detail.component';
import { LogInComponent } from './view/log-in/log-in.component';
import { DataMgComponent } from './view/data-mg/data-mg.component';
import { DataDetailsComponent } from './view/data-details/data-details.component';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { HttpClient, HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { ProjectMgComponent } from './view/project-mg/project-mg.component';
import { AuthInterceptor } from '../utils/auth-interceptor';
import { MessageModule } from './components/message/message.module';
import { AppService, createTranslateLoader } from './app.service'
import { TranslateModule, TranslateLoader } from '@ngx-translate/core'
import { FilterComponent } from './components/filter/filter.component';
import { HomeComponent } from './view/home/home.component'
import { SelectivePreloadingStrategy } from '../utils/selective-preloading-strategy'
import { ShardModule } from 'src/app/shared/shard/shard.module';
import { UploadFileComponent } from './components/upload-file/upload-file.component';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    SidenavComponent,
    ModelMgComponent,
    SiteConfigComponent,
    UserMgComponent,
    ModelDetailComponent,
    LogInComponent,
    DataMgComponent,
    DataDetailsComponent,
    ProjectMgComponent,
    FilterComponent,
    HomeComponent,
    UploadFileComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ClarityModule,
    BrowserAnimationsModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule,
    MessageModule,
    ShardModule,
    TranslateModule.forRoot({// 配置i8n
      defaultLanguage: 'en',
      loader: {
        provide: TranslateLoader,
        useFactory: createTranslateLoader,
        deps: [HttpClient]
      }
    }),
    NgbModule
  ],
  providers: [
    AppService,
    SelectivePreloadingStrategy,
    [{ provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true }]
  ],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
