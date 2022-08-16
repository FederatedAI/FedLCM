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

import { Component, OnInit, Input, Output, EventEmitter, OnChanges, SimpleChanges } from '@angular/core';
@Component({
  selector: 'app-filter',
  templateUrl: './filter.component.html',
  styleUrls: ['./filter.component.scss']
})
export class FilterComponent implements OnInit, OnChanges {
  constructor() {

  }
  ngOnChanges(changes: SimpleChanges): void {
    this._datalist = this.dataList.map(el => {
      if (typeof (el[this.searchKey]) === 'string') {
        el[this.searchKey] = el[this.searchKey].toLowerCase()
        if (this.searchKey2 !== '') {
          el[this.searchKey2] = el[this.searchKey2].toLowerCase()
        }
        return el
      } else {
        return el
      }
    })
  }
  @Input() left = 'auto'
  @Input() right = '0'
  @Input() top = 'auto'
  @Input() bottom = 'auto'
  @Input() width = 300
  @Input() dataList: any[] = [] // data
  @Input() searchKey: string = ''
  @Input() searchKey2: string = ''
  @Input() placeholder: string = ''
  @Output() filterDataList = new EventEmitter()
  fixDataView() {
    const data = {
      searchValue: this.filterSearchValue,
      eligibleList: this.eligibleList
    }
    this.filterDataList.emit(data)
  }
  public filterSearchValue: string = ''
  public eligibleList: any[] = []
  private _datalist: any[] = []
  private _filterSearchValue = ''
  private storageDataList: any[] = []
  ngOnInit(): void {
    this.storageDataList = this.dataList
  }

  inputHandle() {
    this.eligibleList = []
    this._filterSearchValue = this.filterSearchValue.toLowerCase()
    if (!this.filterSearchValue.trim()) {
      this.fixDataView()
      return
    }
    this._datalist.forEach((el: any, index: number) => {
      if (typeof (el[this.searchKey]) === 'string') {
        if (el[this.searchKey].indexOf(this._filterSearchValue) !== -1 || (this.searchKey2 !== '' && el[this.searchKey2].indexOf(this._filterSearchValue) !== -1)) {
          this.eligibleList.push(this.dataList[index])
        }
      } else {
        if (Array.isArray(el[this.searchKey])) {
          this.isArray(el[this.searchKey], this._filterSearchValue, index)
        } else {

        }
      }
    });
    this.fixDataView()
  }
  isArray(arr: any[], str: string, index: number) {
    arr.forEach(el => {
      if (this.eligibleList.find(el => el === this.dataList[index])) return
      for (const key in el) {
        if (key.toLocaleLowerCase().indexOf(str) !== -1 || el[key].toLocaleLowerCase().indexOf(str) !== -1) {
          this.eligibleList.push(this.dataList[index])
        }
      }
    })
  }
}
