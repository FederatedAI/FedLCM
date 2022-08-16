import {
  ComponentFactoryResolver,
  ComponentRef,
  Type,
  Injector,
  Provider,
  ElementRef,
  ComponentFactory,
  Injectable
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