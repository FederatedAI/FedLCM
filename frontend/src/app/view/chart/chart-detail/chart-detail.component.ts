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
import { ActivatedRoute, Router } from '@angular/router';
import { ChartService } from 'src/app/services/common/chart.service';
import { CHARTTYPE, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-chart-detail',
  templateUrl: './chart-detail.component.html',
  styleUrls: ['./chart-detail.component.scss']
})
export class ChartDetailComponent implements OnInit {

  constructor(private chartservice: ChartService, private router: Router, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.showChartDetail();
  }
  chartType = CHARTTYPE;
  constantGather = constantGather;

  uuid = String(this.route.snapshot.paramMap.get('id'));
  chartDetail: any;
  errorMessage = "Service Error!"
  isShowChartDetailFailed: boolean = false
  isPageLoading: boolean = true
  values: any
  about: any
  valueTemplate: any
  //createCodeMirror is to initial the yaml editor window
  createCodeMirror(id: string, data: string, key: 'values' | 'about' | 'valueTemplate') {
    const yamlHTML = document.getElementById(id) as any
    this[key] = window.CodeMirror.fromTextArea(yamlHTML, {
      value: '',
      mode: 'yaml',
      lineNumbers: true,
      indentUnit: 1,
      lineWrapping: true,
      tabSize: 2,
      readOnly: true
    })
    this[key].setValue(data)

  }
  //showChartDetail is to get the chart detail information
  showChartDetail() {
    this.isPageLoading = true
    this.isShowChartDetailFailed = false
    this.chartservice.getChartDetail(this.uuid)
      .subscribe((data: any) => {
        this.chartDetail = data.data;
        this.createCodeMirror('values', this.chartDetail.values, 'values')
        this.createCodeMirror('about', this.chartDetail.about, 'about')
        this.createCodeMirror('values_template', this.chartDetail.values_template, 'valueTemplate')
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isPageLoading = false
          this.isShowChartDetailFailed = true
        }
      );
  }

}
