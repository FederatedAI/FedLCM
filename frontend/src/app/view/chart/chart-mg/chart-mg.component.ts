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

import { Component, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import '@cds/core/file/register.js';
import { ChartService } from 'src/app/services/common/chart.service';
import { CHARTTYPE, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-chart-mg',
  templateUrl: './chart-mg.component.html',
  styleUrls: ['./chart-mg.component.scss']
})
export class ChartMgComponent implements OnInit {

  selectedChartList: any = [];
  openModal: boolean = false;
  newchartForm = this.fb.group({
    name: [''],
    type: [''],
    description: ['']
  });
  constructor(private fb: FormBuilder, private chartservice: ChartService, private router: Router) {
    this.showChartList();
  }

  ngOnInit(): void {
  }

  chartType = CHARTTYPE;
  constantGather = constantGather;

  // Currently, "Add a new chart" is not supported
  onOpenModal() {
    this.openModal = true;
  }
  // resetModal() {
  //   this.newchartForm.reset();
  //   this.openModal = false;
  // }

  chartlist: any = [];
  errorMessage = "Service Error!"
  isShowChartFailed: boolean = false;
  isPageLoading: boolean = true;
  //showChartList is to get the chart list
  showChartList() {
    this.isPageLoading = true;
    this.isShowChartFailed = false;
    this.chartservice.getChartList()
      .subscribe((data: any) => {
        this.chartlist = data.data;
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isShowChartFailed = true
          this.isPageLoading = false
        });
  }

  //refresh is for refresh button
  refresh() {
    this.showChartList();
  }


}
