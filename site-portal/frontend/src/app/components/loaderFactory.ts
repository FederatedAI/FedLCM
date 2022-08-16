import {
  ComponentFactoryResolver,
  Injector,
  Injectable
} from '@angular/core';
import { ComponentLoader } from './components-loader';

@Injectable()
export class ComponentLoaderFactory {
  constructor(private _injector: Injector,
    private _cfr: ComponentFactoryResolver) {

  }

  create<T>(): ComponentLoader<T> {
    return new ComponentLoader(this._cfr, this._injector);
  }
}