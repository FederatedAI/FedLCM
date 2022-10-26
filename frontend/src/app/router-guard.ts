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

import { Injectable } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot, Router } from '@angular/router';
import { AuthService } from './services/common/auth.service';

@Injectable()

export class RouterGuard implements CanActivate {
  constructor(private router: Router, private authService: AuthService) {
  }
  experimentEnabled = false
  public async canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> {
    await new Promise((re, rj) => {
      this.authService.getLCMServiceStatus()
        .subscribe((data: any) => {
          // use the value of 'experiment_enabled' to control enabling openfl related content or not
          this.experimentEnabled = data?.experiment_enabled;
          re(this.experimentEnabled)
        })
    })
    return this.experimentEnabled
  }
}