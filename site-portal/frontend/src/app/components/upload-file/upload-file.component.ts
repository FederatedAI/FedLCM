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

import { Component, OnInit, Input, ViewChild, ElementRef,OnDestroy } from '@angular/core';
import '@cds/core/icon/register.js';
import { folderIcon, ClarityIcons } from '@cds/core/icon';
ClarityIcons.addIcons(folderIcon);
@Component({
  selector: 'app-upload-file',
  templateUrl: './upload-file.component.html',
  styleUrls: ['./upload-file.component.css']
})

//this component is for uploading file
export class UploadFileComponent implements OnInit,OnDestroy {
  constructor() { }
  ngOnDestroy(): void {
    this.fileName = 'BROWSE'
    this.file = ''
    this.fileinput.nativeElement.value = null
  }
  maxSizeFlag = false
  fileType = false
  isUploaded = false
  get size () {
    return this.maxSize + 'MB'
  }
  format = 'csv'
  file: any = ''
  fileName = 'BROWSE'
  @Input() fileList: string[] = []
  @Input() progress = 0
  @Input() maxSize = 500
  @ViewChild('file') fileinput!: ElementRef
  ngOnInit(): void {
  }
  uploadFile(event:Event) {
    this.maxSizeFlag = false
    this.fileType = false
    this.file = (event.target as HTMLInputElement)?.files?.[0];
    this.fileName = this.file.name
    // limit file size of 500m
    if (this.file?.size) {
      const size = this.file.size
      const sizeM = size / 1024 / 1024
      if (sizeM < this.maxSize) {
        // validate file format
        if (this.isCSVFile(this.fileName)) {
          this.isUploaded = true
        } else {
          this.fileinput.nativeElement.value = null
          this.fileType = true
        }
      } else {
        this.fileinput.nativeElement.value = null
        this.maxSizeFlag = true
      }
    }
  }
  empty () {
    this.fileName = 'BROWSE'
    this.file = ''
    this.fileinput.nativeElement.value = null
  }
  input() {
    this.fileinput.nativeElement.click()
  }
  isCSVFile(fileName:string): boolean {
    const suffix = fileName.slice(-4)
    return suffix === ".csv" 
  }
}
