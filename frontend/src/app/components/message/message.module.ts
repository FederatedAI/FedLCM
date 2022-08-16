import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MessageComponent } from './message.component'
import { ClarityModule } from '@clr/angular'
import { ComponentLoaderFactory } from '../loaderFactory'
import { MessageService } from './message.service'
import { TranslateModule } from '@ngx-translate/core'

@NgModule({
  declarations: [MessageComponent],
  imports: [
    CommonModule,
    ClarityModule,
    TranslateModule
  ],
  providers:[MessageService, ComponentLoaderFactory],
  entryComponents: [MessageComponent],
  exports: [MessageComponent]
})
export class MessageModule { }
