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

import {
  ComponentFactoryResolver,
  ComponentRef,
  Type,
  Injector,
  Provider,
  ElementRef,
  ComponentFactory
} from '@angular/core'
export class ComponentLoader<T> {
  constructor(private _cfr: ComponentFactoryResolver,
    private _injector: Injector) {
  }
  private _componentFactory: ComponentFactory<T> | any
  attch(componentType: Type<T>): ComponentLoader<T> {
    this._componentFactory = this._cfr.resolveComponentFactory<T>(componentType)
    return this
  }
  private _parent: Element | any
  to(parent: string | ElementRef): ComponentLoader<T> {
    if (parent instanceof ElementRef) {
      this._parent = parent.nativeElement
    } else {
      this._parent = document.querySelector(parent)
    }
    return this
  }
  private _providers: Provider[] = []
  provider(provider: Provider) {
    this._providers.push(provider)
  }
  create(opts: {}): ComponentRef<T> {
    const injector = Injector.create({
      providers: this._providers as any[],
      parent: this._injector,
      name: '$msg'
    })
    const componentRef = this._componentFactory.create(injector)
    Object.assign(componentRef.instance, opts)
    if (this._parent) {
      this._parent.appendChild(componentRef.location.nativeElement)
    }
    componentRef.changeDetectorRef.markForCheck();
    componentRef.changeDetectorRef.detectChanges();
    return componentRef;
  }
  remove(ref: ComponentRef<T> | any) {
    if (this._parent) {
      this._parent.removeChild(ref.location.nativeElement)
    }
    ref = null;
  }
}