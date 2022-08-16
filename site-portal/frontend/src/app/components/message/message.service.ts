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

import { Injectable, Injector } from '@angular/core';
import { ComponentLoaderFactory } from '../loaderFactory'
import { ComponentLoader } from '../components-loader'
import { MessageComponent } from './message.component'
@Injectable({
  providedIn: 'root'
})
export class MessageService {
  constructor(
    private _clf: ComponentLoaderFactory,
    private _injector: Injector,
  ) {
    this._loader = this._clf.create<MessageComponent>();
  }
  public ref: any
  private _loader: ComponentLoader<MessageComponent>
  private popUpMessage(t: string, messageContent: string, duration = 10000) {
    this._loader.attch(MessageComponent).to('body')
    const opts = {
      messageType: t,
      messageContent: messageContent
    }
    this.ref = this._loader.create(opts)
    this.ref.changeDetectorRef.markForCheck()
    this.ref.changeDetectorRef.detectChanges()
    setTimeout(() => {
      if (this.ref !== '') {
        this._loader.remove(this.ref)
      }
    }, duration)
  }
  public info(messageContent: string, duration?: number) {
    this.popUpMessage('info', messageContent, duration);
  }
  public success(messageContent: string, duration?: number) {
    this.popUpMessage('success', messageContent, duration);
  }
  public error(messageContent: string, duration?: number) {
    this.popUpMessage('error', messageContent, duration);
  }
  public warning(messageContent: string, duration?: number) {
    this.popUpMessage('warning', messageContent, duration);
  }
  public close() {
    if (this.ref !== '') {
      try {
        this._loader.remove(this.ref)
        this.ref = ''
      } catch (error) {

      }
    }
  }
}
