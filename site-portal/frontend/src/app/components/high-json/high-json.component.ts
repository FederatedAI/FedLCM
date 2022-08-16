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

import { Component, OnInit, ViewEncapsulation, Input, OnChanges, SimpleChanges, Output, EventEmitter } from '@angular/core';
import { isCollapsable, json2html, valueSplice } from 'src/utils/high-json'
import * as $ from 'jquery'
@Component({
  selector: 'app-high-json',
  templateUrl: './high-json.component.html',
  styleUrls: ['./high-json.component.css'],
  encapsulation: ViewEncapsulation.None
})

//This component is for highlighting json content
export class HighJsonComponent implements OnInit, OnChanges {
  private _$: any = {}
  @Input() json: string = ""
  @Input() id = ""
  @Output() callback = new EventEmitter()
  callbackFun(data: string) {
    this.callback.emit(data)
  }
  @Output() callback2 = new EventEmitter()
  callbackFun2(data: string) {
    this.callback.emit(data)
  }
  jsonObj: any = {}
  isCollapsable!: Function
  valueSplice!: Function
  constructor() {
    this._$ = $
    this.valueSplice = valueSplice
    this._$.fn.jsonViewer = function (json: string, options: any) {
      options = Object.assign({}, {
        collapsed: false,
        rootCollapsable: true,
        withQuotes: false,
        withLinks: true
      }, options);
      return this.each(() => {
        // Transform to HTML
        var html = json2html(json, options);
        if (options.rootCollapsable && isCollapsable(json)) {
          html = '<a href class="json-toggle"></a>' + html;
        }
        // Insert HTML in target DOM element
        $(this[0]).html(html);
        $(this[0]).addClass('json-document');
        // Bind click on toggle buttons
        $(this[0]).off('click');
        $(this[0]).on('click', 'a.json-toggle', function () {
          var target = $(this).toggleClass('collapsed').siblings('ul.json-dict, ol.json-array');
          target.toggle();
          if (target.is(':visible')) {
            target.siblings('.json-placeholder').remove();
          } else {
            var count = target.children('li').length;
            var placeholder = count + (count > 1 ? ' items' : ' item');
            target.after('<a href class="json-placeholder">' + placeholder + '</a>');
          }
          return false;
        });
        // Simulate click on toggle button when placeholder is clicked
        $(this[0]).on('click', 'a.json-placeholder', function () {
          $(this[0]).siblings('a.json-toggle').click();
          return false;
        });
        if (options.collapsed == true) {
          // Trigger click to collapse all nodes
          $(this[0]).find('a.json-toggle').click();
        }
      });
    }

  }
  ngOnChanges(changes: SimpleChanges): void {
    this.updateJsonAttr(this.id)
  }
  ngOnInit(): void {
    this.updateJsonAttr(this.id)
  }

  updateJsonAttr(id: string) {
    this.jsonObj = JSON.parse(this.json)
    const obj = JSON.parse(this.json)
    this._$('#' + id).jsonViewer(obj, {})
    const _this = this
    $('#' + id + ' li').click(function (event: any) {
      if (event.target.classList[0] === 'json-string' || event.target.classList[0] === 'json-literal') {
        event.stopPropagation()
        const input = document.createElement('input')
        input.className = 'update'
        let str
        let that: any = {}
        if (event.target.innerText.indexOf('"') !== -1) {
          str = event.target.innerText.replace(/\"/g, '')
          input.value = str
        } else {
          input.value = event.target.innerText
        }
        let liStr:any = this.innerText.split(':')
        that = event.target
        this.appendChild(input)
        input.focus()
        $('input.update').blur(function (event: any) {
          const valueType = +event.target.value
          if (isNaN(valueType)) {
            if (event.target.value === 'true' || event.target.value === 'false') {
              that.innerText = event.target.value
              liStr[1] = event.target.value
            } else {
              that.innerText = '"' + event.target.value + '"'
              liStr[1] = event.target.value
            }
          } else {
            that.innerText = valueType
            liStr[1] = valueType
          }
          // liStr[1] = event.target.value
          _this.valueSplice(liStr, obj)
          
          _this.jsonObj = obj
          event.target.parentElement.removeChild(event.target)
        })
        
      }
    })    
  }
}