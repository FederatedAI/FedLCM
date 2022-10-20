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

import { Component, OnInit, OnDestroy, ViewChild, AfterViewInit, ChangeDetectorRef } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { DataService } from '../../service/data.service';
import { ParticipantListResponse } from '../project-details/project-details.component';
import { ProjectService } from '../../service/project.service';
import { MessageService } from '../../components/message/message.service'
import { ValidatorGroup } from '../../../config/validators'
import Dag from '../../../config/dag-drag'
import DagJson from '../../../config/dag'
import '@cds/core/icon/register.js';
import { plusCircleIcon, ClarityIcons, angleIcon } from '@cds/core/icon';
import { HighJsonComponent } from 'src/app/components/high-json/high-json.component'
ClarityIcons.addIcons(plusCircleIcon, angleIcon);

export interface PartyUser {
  creation_time: string,
  description: string,
  name: string,
  party_id: number,
  status: number,
  uuid: string,
  selected: boolean,
  associated_data: string,
  data_list: any,
  label_column: string
}
export interface PredictModel {
  component_name: string,
  create_time: string,
  job_name: string,
  job_uuid: string,
  model_id: string,
  model_version: string,
  name: string,
  party_id: 0,
  project_name: string,
  project_uuid: string,
  role: string,
  uuid: string,
  participant: PartyUser[],
  selected: boolean,
}
interface HostType {
  name: string
  form: any
}
interface AlgorithmType {
  moduleName: string
  parameters: {
    [key: string]: any
  }
  conditions: {
    possible_input: string[],
    can_be_endpoint: boolean,
  },
  input: {
    data: string[],
    model: string[]
  }
  output: {
    data: string[],
    model: string[]
  }
  count?: number
}
interface GroupType {
  groupName: string,
  modules: AlgorithmType[]
  angle?: boolean
}
interface programModol {
  name: string
  value: boolean | string | number
}
interface optRelationshipsType {
  value: string
  type: string
  relation?: string,
  list: any[]
}
interface SvgModel {
  [key: string]: {
    module: string
    parameters: { [key: string]: any },
    attributes: {
      [key: string]: {
      }
    }
    conditions: {
      output?: {
        data: string[],
        model: string[]
      },
      input?: {
        data: string[]
      },
      relation?: any[]
    },
    attributeType: 'common' | 'diff',
    diffAttribute?: any,
    default?: any
  }
}
interface InputOrOutputTYpe {
  data: string[]
  model: string[]
}
interface Relationship {
  data: optRelationshipsType[],
  model: optRelationshipsType[]
}
interface inputModuleType {
  inputModule: string
  optRelationships: optRelationshipsType[]
  inputTypeList: string[]
}

@Component({
  selector: 'app-job-new',
  templateUrl: './job-new.component.html',
  styleUrls: ['./job-new.component.scss', './job-new.css']
})
export class JobNewComponent implements OnInit, OnDestroy, AfterViewInit {
  @ViewChild('dslJson') dslRef!: HighJsonComponent;
  @ViewChild('alJson') alRef!: HighJsonComponent;

  //drag and dag module
  dropOrJson = true // drop
  algorithmDataSourc: GroupType[] = []
  algorithmList: AlgorithmType[] = []
  dragStorageArea: any[] = []

  // Data used to generate structure diagrams
  svgData: SvgModel = {
    reader_0: {
      module: "Reader",
      attributes: {},
      parameters: {},
      conditions: {
        output: {
          data: ['data'],
          model: []
        }
      },
      attributeType: 'common'
    }
  }

  currentDragObj: any = {};

  // drag mode d3 instance
  dag: any = {}
  // json mode d3 instance
  dagJson: any = {}

  dslJson: string = ''
  confJson: string = ''

  // drag mode The current module can input list
  inputModuleList: inputModuleType[] = [{
    inputModule: '',
    optRelationships: [{
      value: '',
      type: '',
      relation: '',
      list: []
    }],
    inputTypeList: []
  }]
  outputRelationships: any[] = []
  JobAlgorithmType = 0
  addModuleFlag = false
  currentModuleAttrForm: any = {}
  dblModuleName = ''
  // current difference attribute
  diff = -1
  // host group
  hostList: HostType[] = []

  form: FormGroup;
  _$: any
  openModal: boolean = false;
  modalSize = "xl";
  program: programModol[] = []
  name: string = "";
  desc: string = "";
  validationDataPercent: string = "";
  model_name: string = "";
  algorithm: string = "";
  algorithmConfig: string = "";
  dsl: string = "";
  options: boolean = true;

  //job type variable
  psi: boolean = false;
  modeling: boolean = true;
  predict: boolean = false;

  participant: string = "";
  newJobType: string = "modeling";
  dataOptions: any;
  columnOptions: any;
  selected: any;
  errorMessage: string = "";

  routeParams = this.route.snapshot.paramMap;
  projectUUID = String(this.routeParams.get('id'));
  allParticipantList: any = [];
  participantList: PartyUser[] = [];
  isShowParticiapantFailed: boolean = false;
  selfAssociatedDataListResponse: any;
  getParticipantAssociatedDataListIsPending: boolean = false;

  //displayParticipant is displayed data configuration of paticipant excepting 'Self'
  displayParticipant: PartyUser[] = [];
  //displaySelf is displayed data configuration of 'self' paticipant
  displaySelf: PartyUser = {
    creation_time: "",
    description: "",
    name: "",
    party_id: 0,
    status: 0,
    uuid: "",
    selected: true,
    associated_data: "",
    data_list: [],
    label_column: ""
  };
  submitSaveSelection: boolean = false;
  invalidSave: boolean = false;
  noAssociatedData: boolean = false;
  showLocalDataListFailed: boolean = false;
  showAssociatedDataSubmit: boolean = false;
  allAssociatedData: any
  modalErrorMessage: string = ''; selfdatalist: any; self: PartyUser = {
    creation_time: "",
    description: "",
    name: "",
    party_id: 0,
    status: 0,
    uuid: "",
    selected: true,
    associated_data: "",
    data_list: [],
    label_column: ""
  }
  dataColumnResponse: any;
  dataColumn: any;
  jobDetail: any = {
    conf_json: "",
    description: "",
    dsl_json: "",
    initiator_data: {
      data_uuid: "",
      label_name: ""
    },
    name: "",
    other_site_data: [],
    predicting_model_uuid: "",
    project_uuid: "",
    training_algorithm_type: 1,
    training_component_list_to_deploy: [] as string[],
    training_model_name: "",
    training_validation_enabled: true,
    training_validation_percent: 0,
    type: 0
  }
  //predictModel is selected model uuid for prediction job
  predictModel: string = "";
  submitNewJobFailed: boolean = false;
  submitNewJob: boolean = false;
  formvalue: any;
  submitGeneratedFailed: boolean = false;
  submitGenerated: boolean = false;
  isShowModelFailed: boolean = false;
  modelList: any;
  displayPredictParticipant: any = [];
  predictParticipantList: any;
  newPredictParticipantList: PartyUser[] = [];
  // Automatically calculate the width of the artboard in drag mode
  get width() {
    if (this.dag.count > 3) {
      return (this.dag.count - 3) * 200
    }
    return 0
  }
  // Automatically calculate the Height of the artboard in drag mode
  get height() {
    if (this.dag.level > 10) {
      return (this.dag.level - 10) * 90
    }
    return 0
  }
  // Dynamically get the output options of the current module
  get changeOptRelationshipList() {
    const keyList: string[] = []
    if (this.currentDragObj.hasOwnProperty('default')) {
      this.currentDragObj.default.conditions.possible_input.forEach((str: string) => {
        for (const key in this.svgData) {
          if (this.svgData[key].module === str) {
            keyList.push(key)
          }
        }
      })
    } else {
      this.currentDragObj.conditions.possible_input.forEach((str: string) => {
        for (const key in this.svgData) {
          if (this.svgData[key].module === str) {
            keyList.push(key)
          }
        }
      })
    }
    return keyList
  }
  // create job button disabled
  get disabled() {
    return this.inputModuleList.every(item => {
      return item.optRelationships.every(el => el.value !== '' && el.type !== '' && el.relation !== '')
    }) && this.outputRelationships.every(el => el.value !== '' && el.type !== '')
  }
  // Toggle module property sheet representation in drag and drop mode
  get currentDiff() {
    return this.diff
  }
  set currentDiff(value) {
    if (value > -1) {
      // when click the module
      if (this.hostList[value].form && JSON.stringify(this.hostList[value].form) !== '{}') {
        // existing host
        this.currentModuleAttrForm = {}
        for (const key in this.hostList[value].form) {
          if (this.svgData[this.dag.attrForm.moduleName].diffAttribute) {
            const obj = JSON.parse(JSON.stringify(this.svgData[this.dag.attrForm.moduleName].diffAttribute['host_'+value]))

            if (Object.prototype.toString.call(this.hostList[value].form[key]) === '[object Object]') {
              const result = this.hostList[value].form[key].hasOwnProperty('drop_down_box')
              if (result) {
                this.currentModuleAttrForm[key] = obj[key]
              } else {
                this.currentModuleAttrForm[key] = JSON.stringify(obj[key])
              }
            } else if (Object.prototype.toString.call(this.hostList[value].form[key]) === '[object Array]') {
              this.currentModuleAttrForm[key] = JSON.stringify(obj[key])
            } else {
              this.currentModuleAttrForm[key] = obj[key]
            }
          } else {
            if (Object.prototype.toString.call(this.hostList[value].form[key]) === '[object Object]') {
              const result = this.hostList[value].form[key].hasOwnProperty('drop_down_box')
              if (result) {
                this.currentModuleAttrForm[key] = this.hostList[value].form[key]
              } else {
                this.currentModuleAttrForm[key] = JSON.stringify(this.hostList[value].form[key])
              }
            } else if (Object.prototype.toString.call(this.hostList[value].form[key]) === '[object Array]') {
              this.currentModuleAttrForm[key] = JSON.stringify(this.hostList[value].form[key])
            } else {
              this.currentModuleAttrForm[key] = this.hostList[value].form[key]
            }
          }

        }
      } else {
        //dag and drag mode:  the first host
        if (value === 0) {
          this.currentModuleAttrForm = {}
          this.dag.attrForm.attrList.forEach((el: any) => {
            const obj = JSON.parse(JSON.stringify(el))
            if (Object.prototype.toString.call(obj.value) === '[object Object]') {
              const result = obj.value.hasOwnProperty('drop_down_box')
              if (result) {
                this.currentModuleAttrForm[el.name] = obj.value
              } else {
                this.currentModuleAttrForm[el.name] = JSON.stringify(obj.value)
              }

            } else if (Object.prototype.toString.call(obj.value) === '[object Array]') {
              this.currentModuleAttrForm[el.name] = JSON.stringify(obj.value)
            } else {
              this.currentModuleAttrForm[el.name] = obj.value
            }
          });
          this.hostList[0] = {
            name: 'host_0',
            form: JSON.parse(JSON.stringify(this.currentModuleAttrForm))
          }
        }
      }
    } else {// when it is guest
      if (this.dag.attrForm.diffList.length > 0) {
        this.dag.attrForm.diffList.forEach((el: any) => {
          if (el.name === 'guest') {
            this.currentModuleAttrForm = {}
            for (const key in el.form) {
              if (Object.prototype.toString.call(el.form[key]) === '[object Object]') {
                const result = el.form[key].hasOwnProperty('drop_down_box')
                if (result) {
                  this.currentModuleAttrForm[key] = el.form[key]
                } else {
                  this.currentModuleAttrForm[key] = JSON.stringify(el.form[key])
                }
              } else if (Object.prototype.toString.call(el.form[key]) === '[object Array]') {
                this.currentModuleAttrForm[key] = JSON.stringify(el.form[key])
              } else {
                this.currentModuleAttrForm[key] = el.form[key]
              }
            }
          }
        });
      } else {
        if (this.svgData[this.dag.attrForm.moduleName].diffAttribute) {
          this.currentModuleAttrForm = {}
          const obj = JSON.parse(JSON.stringify(this.svgData[this.dag.attrForm.moduleName].diffAttribute.guest))
          for (const key in obj) {
            if (Object.prototype.toString.call(obj[key]) === '[object Object]') {
              const result = obj[key].hasOwnProperty('drop_down_box')
              if (result) {
                this.currentModuleAttrForm[key] = obj[key]
              } else {
                this.currentModuleAttrForm[key] = JSON.stringify(obj[key])
              }

            } else if (Object.prototype.toString.call(obj[key]) === '[object Array]') {
              this.currentModuleAttrForm[key] = JSON.stringify(obj[key])
            } else {
              this.currentModuleAttrForm[key] = obj[key]
            }
          }
        }
      }
    }
    this.diff = value
  }
  // Decide whether to enable different radio buttons according to the number of Data Configuration
  get diffShow() {
    if (this.dag.attrForm.moduleName === 'reader_0') {
      return true
    }
    return this.displayParticipant.filter(el => el.selected).length > 0

  }

  constructor(private fb: FormBuilder, private route: ActivatedRoute, private projectservice: ProjectService, private router: Router, private dataservice: DataService, private msg: MessageService, private cdRef: ChangeDetectorRef) {

    this.form = this.fb.group(
      ValidatorGroup([
        {
          name: 'name',
          type: ['word'],
          max: 20,
          min: 2
        },
        {
          name: 'desc',
          type: ['']
        },
        {
          name: 'validationDataPercent',
          type: ['notRequired', 'zero']
        },
        {
          name: 'newJobType',
          type: ['']
        },
        {
          name: 'model_name',
          type: ['word'],
          max: 20,
          min: 2
        },
        {
          name: 'algorithmConfig',
          type: ['']
        },
        {
          name: 'algorithm',
          type: ['']
        },
        {
          name: 'dsl',
          type: ['']
        },
        {
          name: 'predictModel',
          type: ['']
        },
      ])
    );
  }


  ngOnInit(): void {
    this.showParticipantList();
    this.getDataSourc()
    this.dag = new Dag(this.svgData, '#svg-canvas-drop', this.moduleAttrSave, this)
    this.startDag('{}', '{}')
    this.dag.Generate()
    this.program = []
    setTimeout(() => {
      this.dag.Draw()
      this.dropOrJson = false
    })
  }
  ngOnDestroy(): void {
    //leave close message
    this.msg.close()
  }

  ngAfterViewInit(): void {
    this.cdRef.detectChanges();
  }
  addOptRelationship(item: inputModuleType) {
    item.optRelationships.push({
      value: item.inputModule,
      type: '',
      relation: '',
      list: []
    })
  }
  addInputModule() {
    this.inputModuleList.push({
      inputModule: '',
      optRelationships: [{
        value: '',
        type: '',
        relation: '',
        list: []
      }],
      inputTypeList: []
    })
  }
  deleteItem(index: number, item: inputModuleType | 'inputModuleList') {
    if (item === 'inputModuleList') {
      this[item].splice(index, 1)
    } else {
      item.optRelationships.splice(index, 1)
    }
  }
  // Drag and drop an algorithm module to trigger a function
  dragHandler(e: any) {
    // get drag and dag object
    this.currentDragObj = e.previousContainer.data[e.previousIndex] as AlgorithmType
    this.inputModuleList = [{
      inputModule: '',
      optRelationships: [{
        value: '',
        type: '',
        relation: '',
        list: []
      }],
      inputTypeList: []
    }]
    if (e.previousContainer.data[e.previousIndex].moduleName === 'HomoLR') {
      this.JobAlgorithmType = 1
    } else if (e.previousContainer.data[e.previousIndex].moduleName === 'HomoSecureboost') {
      this.JobAlgorithmType = 2
    } else if (e.previousContainer.data[e.previousIndex].moduleName === 'HeteroLR') {
      this.JobAlgorithmType = 3
    } else if (e.previousContainer.data[e.previousIndex].moduleName === 'HeteroSecureBoost') {
      this.JobAlgorithmType = 4
    }
    // true Indicates a new drag and drop
    this.bulletFrame(true, '')
  }
    // handle pop-up modal
  bulletFrame(bool: boolean, str: string) {
    this.dblModuleName = str
    if (bool) {
      // get the available input list
      this.outputRelationships = []
      //Get the output item of the current drag module
      this.getOutputListHandler(this.currentDragObj.output)
    }
    else {
      this.inputModuleList = []
      const input = this.svgData[str].conditions.input
      if (input) {
        input.data.forEach(el => {
          this.inputModuleList.push({
            inputModule: el,
            optRelationships: [],
            inputTypeList: []
          })
        })
      }
      this.currentDragObj = this.svgData[str]
      this.currentDragObj.moduleName = this.currentDragObj.module
      this.inputModuleList.forEach(item => {
        if (this.svgData[item.inputModule]) {
          const relation = this.svgData[str].conditions.relation
          const moudel = this.svgData[item.inputModule].default
          if (moudel) {
            if (moudel.output.model.length > 0) {
              item.inputTypeList = [...moudel.output.model, ...moudel.output.data]
            } else {
              item.inputTypeList = [...moudel.output.data]
            }
          } else {
            item.inputTypeList = ['data']
          }
          if (relation) {
            for (const key in relation) {
              if (key === item.inputModule) {
                relation[key].list.forEach((el: any) => {
                  if (el.value === 'model') {
                    item.optRelationships.push({
                      value: item.inputModule,
                      type: el.key,
                      relation: el.value,
                      list: this.currentDragObj.default.input.model
                    })
                  } else {
                    item.optRelationships.push({
                      value: item.inputModule,
                      type: el.key,
                      relation: el.value,
                      list: this.currentDragObj.default.input.data
                    })
                  }
                })
              }
            }
          }
          this.outputRelationships = []
        } else {
          item.optRelationships[0] = {
            value: '',
            relation: '',
            type: '',
            list: []
          }
          item.inputModule = ''
        }
      })
      // get the available output list
      this.getOutputListHandler(this.currentDragObj.conditions)
    }
    this.addModuleFlag = true
  }
  currentModel(item: inputModuleType) {
    let moudel: any = {}
    for (const data in this.svgData) {
      if (data === item.inputModule) {
        moudel = this.svgData[data]
      }
    }
    if (moudel) {
      if (moudel.conditions.output.model.length > 0) {
        item.inputTypeList = ['', ...moudel.conditions.output.model, ...moudel.conditions.output.data]
      } else {
        item.inputTypeList = ['', ...moudel.conditions.output.data]
      }
    } else {
      item.inputTypeList = ['', 'data']
    }
  }
  optChange(item: inputModuleType, opt: optRelationshipsType, type?: string) {
    opt.value = item.inputModule
    if (this.currentDragObj.input) {
      if (type === 'model') {
        opt.list = this.currentDragObj.input.model
      } else {
        opt.list = this.currentDragObj.input.data
      }
    } else {
      if (type === 'model') {
        opt.list = this.currentDragObj.default.input.model
      } else {
        opt.list = this.currentDragObj.default.input.data
      }
    }

  }

  diffGuestOrHostChangeHandler(num:number) {
    // this.currentDiff = num;
    console.log('num', num);
    console.log('algorithmList', this.algorithmList);
    this.algorithmList.forEach(el => {
      if (el.moduleName === this.dag.attrForm.moduleName) {
        this.currentModuleAttrForm = el.parameters
      }
    })
    console.log('svgData', this.svgData);
    console.log('currentModuleAttrForm', this.currentModuleAttrForm);

  }
  getOutputListHandler(output: any) {
    if (output.model) {
      output.data.forEach((el: string) => {
        this.outputRelationships.push({
          value: el,
          type: 'data'
        })
      });
      output.model.forEach((el: string) => {
        this.outputRelationships.push({
          value: el,
          type: 'model'
        })
      });
    } else {
      output.data.forEach((el: string) => {
        this.outputRelationships.push({
          value: el,
          type: 'data'
        })
      });
    }
  }


  // re-render d3 legend
  canvasRedraw() {
    this.dag.dsl = this.svgData
    this.dag.Generate()
    this.dag.Draw()
    this.program = this.dag.my
    this.program.forEach(el => {
      if (el.name.indexOf('Evaluation') === -1 && el.name.indexOf('HomoDataSplit') === -1 && el.name.indexOf('HeteroDataSplit') === -1) {
        el.value = true
      } else {
        el.value = false
      }
      return el
    })
  }

  // darg mode zoom
  zoom(num: number) {
    if (num === -0.1 && String(this.dag.zoomMultiples).slice(0, 3) === '0.1') {
      return
    }
    this.dag.zoomMultiples += num
    this.dag.Draw()
  }

  // Drag-and-drop mode determines module additions
  addAlgorithmHandler(type: string) {
    const svgObj: any = {
      module: this.currentDragObj.moduleName,
      attributes: this.currentDragObj.parameters,
      attributeType: type,
      conditions: {
        input: {
          data: [],
          model: []
        },
        output: {
          data: [],
          model: []
        },
        relation: {}
      },
      default: { ...this.currentDragObj }
    }
    if (type !== 'common') {
      svgObj.attributeType = 'diff'
      svgObj.diffAttribute = this.currentDragObj.diffAttribute
    }
    this.inputModuleList.forEach(item => {
      item.optRelationships.forEach(el => {
        const arr = svgObj.conditions.input.data
        if (svgObj.conditions.relation.hasOwnProperty(item.inputModule)) {
          svgObj.conditions.relation[item.inputModule].list.push({
            key: el.type,
            value: el.relation
          })
        } else {
          svgObj.conditions.relation[item.inputModule] = {
            name: item.inputModule,
            list: [{
              key: el.type,
              value: el.relation
            }]
          }
        }
        if (!arr.find((el: string) => el === item.inputModule)) {
          arr.push(item.inputModule)
        }
      })
    })
    this.outputRelationships.forEach(el => {
      svgObj.conditions.output.data.push(el.value)
    })
    return svgObj
  }
  // Drag and drop modules to confirm adding or modifying functions
  sureOptRelationship() {
    this.dag.attrForm.moduleName = ''
    this.dag.attrForm.attrList = []
    this.currentModuleAttrForm = {}
    let svgObj: any = {}
    if (this.svgData.hasOwnProperty(this.dblModuleName)) {// modify
      svgObj = this.addAlgorithmHandler(this.currentDragObj.attributeType)
      this.svgData[this.dblModuleName] = svgObj
      this.dag.attrForm.moduleName = this.dblModuleName
      this.canvasRedraw()
      for (const key in this.currentDragObj.default.parameters) {
        this.dag.attrForm.attrList.push({
          name: key,
          value: this.currentDragObj.default.parameters[key],
        })
        this.currentModuleAttrForm[key] = this.currentDragObj.default.parameters[key]
      }
    } else {// new add
      svgObj = this.addAlgorithmHandler('common')
      this.svgData[this.currentDragObj.moduleName + '_' + this.currentDragObj.count] = svgObj
      this.dag.attrForm.moduleName = this.currentDragObj.moduleName + '_' + this.currentDragObj.count
      this.algorithmList.forEach((el: any) => {
        if (el.moduleName === this.currentDragObj.moduleName) {
          el.count++
        }
      });
      this.canvasRedraw()
      // get attribute list, duplicate corresponding stroage object
      for (const key in this.currentDragObj.parameters) {
        this.dag.attrForm.attrList.push({
          name: key,
          value: this.currentDragObj.parameters[key],
        })
        if (Object.prototype.toString.call(this.currentDragObj.parameters[key]) === '[object Object]') {
          const result = this.currentDragObj.parameters[key].hasOwnProperty('drop_down_box')
          if (result) {
            this.currentModuleAttrForm[key] = this.currentDragObj.parameters[key]
          } else {
            this.currentModuleAttrForm[key] = JSON.stringify(this.currentDragObj.parameters[key])
          }
        } else if (Object.prototype.toString.call(this.currentDragObj.parameters[key]) === '[object Array]') {
          this.currentModuleAttrForm[key] = JSON.stringify(this.currentDragObj.parameters[key])
        } else {
          this.currentModuleAttrForm[key] = this.currentDragObj.parameters[key]
        }
      }
    }
    this.addModuleFlag = false
    this.outputRelationships = []
    this.dag.attrForm.options = 'common'
    this.diff = -1
    this.resetHostList()
  }

  resetHostList() {
    this.hostList.forEach(el => el.form = {})
  }

  // Drag-and-drop mode cancels module addition
  cancelOptRelationship() {
    this.addModuleFlag = false
    this.outputRelationships = []
  }

  // Expand Algorithm Grouping handler
  openAlgorithmGroup(i: number) {
    this.algorithmDataSourc[i].angle = !this.algorithmDataSourc[i].angle
  }

  // Save the properties of the current module to svgData via the form
  moduleAttrSave(item: { [key: string]: any }, that: JobNewComponent) {
    that.currentModuleAttrForm = {}
    if (item.options === 'common') {
      item.attrList.forEach((el: any) => {
        that.currentModuleAttrForm[el.name] = el.value
      });
    } else {
      that.resetHostList()
      // reset default
      that.diff = -1
      item.diffList.forEach((el: any, index: number) => {
        if (el.name === 'guest') {
          that.currentModuleAttrForm = el.form
        } else {
          if (index>0) {
            that.hostList[index - 1] = {
              name: el.name,
              form: el.form
            }
          } else {
            that.hostList[index] = {
              name: el.name,
              form: el.form
            }
          }
        }
      });
      // reset default
      that.currentDiff = -1
    }
  }

  // switch current module property type
  changeCommonOrDiffRadio() {
    if (this.dag.attrForm.options === 'common') {
      if (this.svgData.hasOwnProperty(this.dag.attrForm.moduleName)) {
        this.currentModuleAttrForm = this.svgData[this.dag.attrForm.moduleName].attributes
        for (const attr in this.svgData[this.dag.attrForm.moduleName].attributes) {
          if (Object.prototype.toString.call(this.svgData[this.dag.attrForm.moduleName].attributes[attr]) === '[object Object]'
            || Object.prototype.toString.call(this.svgData[this.dag.attrForm.moduleName].attributes[attr]) === '[object Array]') {
              const result = this.svgData[this.dag.attrForm.moduleName].attributes[attr].hasOwnProperty('drop_down_box')
              if (result) {
                this.currentModuleAttrForm[attr] = this.svgData[this.dag.attrForm.moduleName].attributes[attr]
              } else {
                this.currentModuleAttrForm[attr] = JSON.stringify(this.svgData[this.dag.attrForm.moduleName].attributes[attr])
              }

          } else {
            this.currentModuleAttrForm[attr] = this.svgData[this.dag.attrForm.moduleName].attributes[attr]
          }
        }
      } else {
        this.dag.attrForm.attrList.forEach((el: any) => {
          this.currentModuleAttrForm[el.name] = JSON.parse(JSON.stringify(el))
        });
      }
    } else {
      this.currentDiff = -1
    }
  }

  // Click the module in the d3 legend to save the clicked module attribute to the form
  saveCurrentModuleAttr() {
    const obj: { [key: string]: any } = {}
    const json = JSON.stringify(this.currentModuleAttrForm)
    const currentModuleAttrForm = JSON.parse(json)
    this.dag.attrForm.attrList.forEach((el: any) => {
      obj[el.name] = currentModuleAttrForm[el.name]
    });

    if (this.dag.attrForm.options === 'common') {
      for (const key in this.svgData) {
        if (key === this.dag.attrForm.moduleName) {
          this.svgData[key].attributes = obj
          this.svgData[key].attributeType = this.dag.attrForm.options
        }
      }
    } else { // diffrence
      if (this.currentDiff === -1) {
        for (const key in this.svgData) {
          if (key === this.dag.attrForm.moduleName) {
            if (this.svgData[this.dag.attrForm.moduleName].diffAttribute) {
              const obj2: { [key: string]: any } = this.svgData[this.dag.attrForm.moduleName].diffAttribute['host_0']
              this.svgData[key].diffAttribute = { guest: obj, 'host_0': obj2 }
            } else {
              this.svgData[key].diffAttribute = { guest: obj, 'host_0': obj }
            }
            this.dag.attrForm.diffList.push({
              name: 'guest',
              form: obj
            })
            this.svgData[key].attributeType = this.dag.attrForm.options
          }
        }
      } else {// host
        for (const key in this.svgData) {
          if (key === this.dag.attrForm.moduleName) {
            const str = 'host_' + this.currentDiff
            const index = this.hostList.findIndex((el: any) => el.name === str)
            if (index !== -1) {
              this.hostList.splice(index, 1, { name: str, form: obj })
            } else {
              this.hostList.push({
                name: str,
                form: obj
              })
            }
            if (this.svgData[key].diffAttribute) {// first add
              this.svgData[key].diffAttribute[str] = obj
              this.svgData[key].diffAttribute.guest = this.svgData[key].attributes
              this.svgData[key].attributeType = this.dag.attrForm.options
            } else {
              this.svgData[key].diffAttribute = {}
              this.svgData[key].diffAttribute[str] = obj
              this.svgData[key].diffAttribute.guest = this.svgData[key].attributes
              this.svgData[key].attributeType = this.dag.attrForm.options
            }
          }
        }
      }
    }
    this.dag.dsl = this.svgData
  }

  // switch drop copy
  switchDropOrCopy(bool: boolean) {
    this.dropOrJson = bool
    this.program = []
    this.dsl = ''
    this.algorithmConfig = ''


    // Judge reader_0 type according to the number of data config
    if(this.displayParticipant.filter(el => el.selected).length > 0) {
      this.svgData['reader_0'].attributeType = 'diff'
      if (this.dag.attrForm.moduleName === 'reader_0') {
        this.dag.attrForm.options = 'diff'
      }
    } else {
      this.svgData['reader_0'].attributeType = 'common'
      if (this.dag.attrForm.moduleName === 'reader_0') {
        this.dag.attrForm.options = 'common'
      }

    }
    if (!this.dropOrJson) {
      this.svgData = {
        reader_0: {
          module: "Reader",
          attributes: {},
          parameters: {},
          conditions: {
            output: {
              data: ['data'],
              model: []
            }
          },
          attributeType: 'common'
        }
      }
      this.canvasRedraw()
    } else {
      this.dagJson.d3.selectAll("#svg-canvas > *").remove();
      this.algorithmList.forEach(el => {
        el.count = 0
      })
    }
  }

  // get Algorithm Data Source
  getDataSourc() {
    this.projectservice.getAlgorithmData().subscribe(
      (data: any) => {
          const getData = JSON.parse(data.data)
          this.algorithmDataSourc = getData.map((el: any) => {
          el.angle = true
          el.modules.forEach((item: any, i: number) => {
            item.count = 0
            for (const key in item.parameters) {
              if(Object.prototype.toString.call(item.parameters[key]) === '[object Object]') {
                if (item.parameters[key].hasOwnProperty('drop_down_box')) {
                  item.parameters[key] = {
                    drop_down_box: item.parameters[key].drop_down_box,
                    value: item.parameters[key].drop_down_box[0]
                  }
                }
              }
            }
            this.algorithmList.push(item)
          });
          return el
        })
      }
    )
  }
  // get dsl conf
  getDslConf() {
    this.saveNewJob(false);
    this.submitGenerated = true;
    this.submitGeneratedFailed = this.checkSubmitValid(false);
    if (!this.submitGeneratedFailed) {
      const data = {
        jobDetail: this.jobDetail,
        interactive: this.svgData
      }
      // Determine the number of data config selected to determine the render data structure
      if (this.jobDetail.other_site_data.length > 0) {
          this.svgData['reader_0'].diffAttribute = {
            guest: {}
          }
        this.jobDetail.other_site_data.forEach((el:any, index:number)=> {
          this.svgData['reader_0'].diffAttribute['host_'+index]
        });
      } else {
        this.svgData['reader_0'].diffAttribute = {}
        this.svgData['reader_0'].attributeType = 'common'
      }


      // this.hostList.forEach(el => {
      //   this.svgData['reader_0'].diffAttribute[el.name] = {}
      // })

      const reqData = JSON.parse(JSON.stringify(this.processingStructure(data)))
      for (const key in reqData.interactive) {
        for (const key2 in reqData.interactive[key].attributes) {
          if (reqData.interactive[key].attributes[key2].hasOwnProperty('drop_down_box')) {
            reqData.interactive[key].attributes[key2] = reqData.interactive[key].attributes[key2].value
          }
        }
      }
      for (const key in reqData.reqData) {
          for (const key3 in reqData.reqData[key].commonAttributes) {
            if (reqData.reqData[key].commonAttributes[key3].hasOwnProperty('drop_down_box')) {
              reqData.reqData[key].commonAttributes[key3] = reqData.reqData[key].commonAttributes[key3].value
            }
          }
          for (const key2 in reqData.reqData[key].diffAttributes) {
            for (const key4 in reqData.reqData[key].diffAttributes[key2]) {
              if (reqData.reqData[key].diffAttributes[key2][key4].hasOwnProperty('drop_down_box')) {
                reqData.reqData[key].diffAttributes[key2][key4] = reqData.reqData[key].diffAttributes[key2][key4].value
              }
            }
          }
      }
      // dsl
      this.projectservice.getDslAndConf(reqData, 'generateDslFromDag').subscribe(
        (data: any) => {
          this.dsl = data.data
          this.submitGeneratedFailed = false
        },
        err => {
          this.submitGeneratedFailed = true
        }
      )
      // conf
      this.projectservice.getDslAndConf(reqData, 'generateConfFromDag').subscribe(
        (data: any) => {
          this.algorithmConfig = data.data
          this.submitGeneratedFailed = false
        },
        err => {
          this.submitGeneratedFailed = true
        }
      )
    }
  }
  // Processing data structure
  processingStructure(data: any) {
    const interactive = JSON.parse(JSON.stringify(data.interactive))
    const reqData: any = {}
    for (const key in interactive) {
      if (interactive[key].attributeType === 'common') {
        for (const attr in interactive[key].attributes) {
          interactive[key].attributes[attr] = this.jsonToObj(interactive[key].attributes[attr])
        }
        reqData[key] = {
          attributeType: interactive[key].attributeType,
          commonAttributes: interactive[key].attributes,
          diffAttributes: {},
          conditions: {
            input: {
              data: {}
            },
            output: {
              data: [],
              model: []
            }
          },
          module: interactive[key].module
        }
        if (interactive[key].conditions.input) {
          const input = interactive[key].conditions.input.data
          const relation = interactive[key].conditions.relation
          input.forEach((i: string) => {
            for (const reKey in relation) {
              if (reKey === i) {
                relation[reKey].list.forEach((el: any) => {
                  reqData[key].conditions.input.data[el.key] = [i + '.' + el.value]
                });
              }
            }
          });
        }
        const output = interactive[key].conditions.output
        output.data.forEach((el: string) => {
          if (el === 'model') {
            reqData[key].conditions.output.model.push(el)
          } else {
            reqData[key].conditions.output.data.push(el)
          }
        });
        if (reqData[key].conditions.output.model.length == 0) {
          delete reqData[key].conditions.output.model
        }
      } else {
        if (key !== 'reader_0') {
          const diffAttribute = interactive[key].diffAttribute
          for (const attr in diffAttribute) {
            for (const key in diffAttribute[attr]) {
              diffAttribute[attr][key] = this.jsonToObj(diffAttribute[attr][key])
            }
          }
          reqData[key] = {
            attributeType: interactive[key].attributeType,
            commonAttributes: {},
            diffAttributes: interactive[key].diffAttribute ? interactive[key].diffAttribute : {},
            conditions: {
              input: {
                data: {}
              },
              output: {
                data: [],
                model: []
              }
            },
            module: interactive[key].module
          }
          if (interactive[key].conditions.input) {
            const input = interactive[key].conditions.input.data
            const relation = interactive[key].conditions.relation
            input.forEach((i: string) => {
              for (const reKey in relation) {
                if (reKey === i) {
                  relation[reKey].list.forEach((el: any) => {
                    reqData[key].conditions.input.data[el.key] = [i + '.' + el.value]
                  });
                }
              }
            });
          }
          const output = interactive[key].conditions.output
          output.data.forEach((el: string) => {
            if (el === 'model') {
              reqData[key].conditions.output.model.push(el)
            } else {
              reqData[key].conditions.output.data.push(el)
            }
          });
          if (reqData[key].conditions.output.model.length == 0) {
            delete reqData[key].conditions.output.model
          }
        } else {
          reqData[key] = {
            attributeType: "diff",
            commonAttributes: {},
            diffAttributes: interactive[key].diffAttribute,
            conditions: {
              output: {
                data: ["data"]
              }
            },
            module: "Reader"
          }
        }
      }
    }
    data.reqData = reqData
    return data
  }
  // JSON string processing
  jsonToObj(value: any) {
    if (typeof (value) !== 'string') {
      return value
    } else {
      try {
        return JSON.parse(value)
      } catch (error) {
        return value
      }
    }

  }
  // set config
  setJsonOrDrag() {
    if (this.dropOrJson) {
      this.getDslConf()
    } else {
      if (this.dsl.length > 0 && this.algorithmConfig.length > 0) {
        this.startDag(JSON.stringify(this.dslRef.jsonObj), JSON.stringify(this.alRef.jsonObj))
      }
    }
  }
  // interactive validator
  interactiveValidator(): boolean {
    const keyArr = Object.values(this.svgData)
    this.submitGeneratedFailed = false;
    this.submitGenerated = false
    if (!keyArr.find(el => el.module === 'HomoLR' || el.module === 'HomoSecureboost' || el.module === 'HeteroLR' || el.module === 'HeteroSecureBoost')) {
      this.errorMessage = 'Homolr or homosecureboost or heterolr or heterosecureboost module is missing'
      this.submitGeneratedFailed = true;
      this.submitGenerated = true
      return false
    }
    if (!keyArr.find(el => el.module === 'Evaluation')) {
      this.errorMessage = 'Missing evaluation module'
      this.submitGeneratedFailed = true;
      this.submitGenerated = true
      return false
    }
    return true
  }

  //selectionOnChange is triggered when the selection of job type is changed
  selectionOnChange(val: any) {
    this.resetSelf();
    this.resetParticipant();
    this.resetConfig();
    if (val == "psi") {
      this.psi = true;
      this.modeling = false;
      this.predict = false;
    }
    if (val == "modeling") {
      this.psi = false;
      this.modeling = true;
      this.predict = false;
    }
    if (val == "predict") {
      this.psi = false;
      this.modeling = false;
      this.predict = true;
    }
    if (this.predict) {
      this.showModelList();
    } else {
      this.hostList = []
      this.displayParticipant = JSON.parse(JSON.stringify(this.participantList));
      const selectList =  this.displayParticipant.filter(el => el.selected)
      for (let i = 0; i < selectList.length; i++) {
        this.hostList.push({
          name: 'host_' + i,
          form: {} as any
        })
      }
    }
  }

  //resetSelf is to reset the data selection of current site itself
  resetSelf() {
    this.self.associated_data = "";
    this.self.data_list = [];
    this.self.label_column = "";
    this.displaySelf = JSON.parse(JSON.stringify(this.self));
    this.displayPredictParticipant = [];
  }

  //resetParticipant is to reset the data selection of selected participant
  resetParticipant() {
    this.participantList = [];
    for (let participant of this.allParticipantList) {
      if (participant.is_current_site) {
        this.self.creation_time = participant.creation_time;
        this.self.description = participant.description,
          this.self.name = participant.name,
          this.self.party_id = participant.party_id,
          this.self.status = participant.status,
          this.self.uuid = participant.uuid
      } else {
        if (participant.status === 1 || participant.status === 3) {
          const party: PartyUser =
          {
            creation_time: participant.creation_time,
            description: participant.description,
            name: participant.name,
            party_id: participant.party_id,
            status: participant.status,
            uuid: participant.uuid,
            selected: false,
            associated_data: "",
            data_list: [],
            label_column: ""
          };
          this.participantList.push(party);
        }
      }
    }
    this.hostList = []
    //display the selected participant and data
    this.displayParticipant = JSON.parse(JSON.stringify(this.participantList));
    const selectList =  this.displayParticipant.filter(el => el.selected)
    for (let i = 0; i < selectList.length; i++) {
      this.hostList.push({
        name: 'host_' + i,
        form: {} as any
      })
    }
  }

  //resetConfig is to reset all configuration of job, which will be triggered when the selection of job type is changed
  resetConfig() {
    this.submitGenerated = false;
    this.invalidSave = false;
    this.submitNewJobFailed = false;
    this.submitNewJob = false;
    this.submitSaveSelection = false;
    this.validationDataPercent = "";
    this.model_name = "";
    this.algorithm = "";
    this.algorithmConfig = "";
    this.dsl = "";
  }

  //showParticipantList is to get all participants
  showParticipantList() {
    this.projectservice.getParticipantList(this.projectUUID, false)
      .subscribe((data: ParticipantListResponse) => {
        this.allParticipantList = data.data;
        this.initParticipantList();
      },
        err => {
          this.isShowParticiapantFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

  //initParticipantList is to init the participant list
  initParticipantList() {
    this.participantList = [];
    for (let participant of this.allParticipantList) {
      if (participant.is_current_site) {
        this.self.creation_time = participant.creation_time;
        this.self.description = participant.description,
          this.self.name = participant.name,
          this.self.party_id = participant.party_id,
          this.self.status = participant.status,
          this.self.uuid = participant.uuid
      } else {
        if (participant.status === 1 || participant.status === 3) {
          const party: PartyUser =
          {
            creation_time: participant.creation_time,
            description: participant.description,
            name: participant.name,
            party_id: participant.party_id,
            status: participant.status,
            uuid: participant.uuid,
            selected: false,
            associated_data: "",
            data_list: [],
            label_column: ""
          };
          this.participantList.push(party);
        }
      }
    }
    this.getSelfAssociatedDataList();
    this.displaySelf = JSON.parse(JSON.stringify(this.self));
    this.hostList = []
    this.displayParticipant = JSON.parse(JSON.stringify(this.participantList));
    const selectList =  this.displayParticipant.filter(el => el.selected)
    for (let i = 0; i < selectList.length; i++) {
      this.hostList.push({
        name: 'host_' + i,
        form: {} as any
      })
    }
  }
  //getSelfAssociatedDataList is get current site associated data list
  getSelfAssociatedDataList() {
    this.projectservice.getParticipantAssociatedDataList(this.projectUUID, this.self.uuid)
      .subscribe(data => {
        this.selfAssociatedDataListResponse = data;
        this.selfdatalist = this.selfAssociatedDataListResponse.data;
      });
  }

  //getParticipantAssociatedDataList
  getParticipantAssociatedDataList(party: PartyUser, selected: boolean, party_uuid: string) {
    this.getParticipantAssociatedDataListIsPending = true;
    if (selected) {
      this.projectservice.getParticipantAssociatedDataList(this.projectUUID, party_uuid)
        .subscribe(data => {
          party.data_list = data.data;
          this.getParticipantAssociatedDataListIsPending = false;
        },
          err => {
            this.getParticipantAssociatedDataListIsPending = false;
            this.errorMessage = err.error.message;
          });
    }
  }

  //saveSelection is to save the current data configuration of participant and 'Self' and close the modal
  saveSelection() {
    this.submitSaveSelection = true;
    this.invalidSave = false;
    this.checkSelection(this.self, this.participantList);
    if (this.invalidSave) {
      this.modalErrorMessage = "Please select all required field";
      return;
    }
    this.displaySelf = JSON.parse(JSON.stringify(this.self));
    this.hostList = []
    this.displayParticipant = JSON.parse(JSON.stringify(this.participantList));
    const selectList =  this.displayParticipant.filter(el => el.selected)
    for (let i = 0; i < selectList.length; i++) {
      this.hostList.push({
        name: 'host_' + i,
        form: {} as any
      })
    }
    this.displayPredictParticipant = JSON.parse(JSON.stringify(this.newPredictParticipantList));
    this.openModal = false;
  }

  //resetSelection is to reset the data confiuration of participant to last saved setting
  resetSelection(modalStatus: boolean) {
    this.self = JSON.parse(JSON.stringify(this.displaySelf));
    this.participantList = JSON.parse(JSON.stringify(this.displayParticipant));
    this.newPredictParticipantList = JSON.parse(JSON.stringify(this.displayPredictParticipant));
    this.openModal = modalStatus;
    this.invalidSave = false;
    this.submitSaveSelection = false;
  }

  //checkSelection is to validate the selection of data confiuration of participants when user trys to save
  checkSelection(self: PartyUser, participantList: PartyUser[]) {
    if (self.associated_data === "") {
      this.invalidSave = true;
    } else {
      if (this.modeling) {
        if (self.label_column === "") this.invalidSave = true;
      }
    }
    if (this.predict) {
      for (let party of this.newPredictParticipantList) {
        if (party.associated_data === "") this.invalidSave = true;
      }
    } else {
      for (let party of participantList) {
        if (party.selected) {
          if (party.associated_data === "") this.invalidSave = true;
        }
      }
    }
  }

  //checkAssociatedData is to check if there is any associated data availble in current project and alert user if there is not
  checkAssociatedData() {
    this.openModal = true;
    this.showAssociatedDataSubmit = true;
    this.projectservice.getAssociatedDataList(this.projectUUID)
      .subscribe(data => {
        this.noAssociatedData = false;
        this.allAssociatedData = data.data;
        if (this.allAssociatedData.length === 0) {
          this.noAssociatedData = true;
          this.showLocalDataListFailed = true;
          this.modalErrorMessage = "No associated data available."
        }
      },
        err => {
          this.showLocalDataListFailed = true;
          this.modalErrorMessage = err.error.message;
        }
      );
  }

  //redirectToData is routing to Data management page in current project
  redirectToData() {
    this.router.navigate(['project-management', 'project-detail', this.projectUUID, 'data']);
  }

  //showDataColumn is to get data column of dataset
  showDataColumn(val: string) {
    let data_id = val.split('+')[1];
    this.dataservice.getDataColumn(data_id)
      .subscribe(data => {
        this.dataColumnResponse = data;
        this.dataColumn = this.dataColumnResponse.data;
      });
  }

  //saveNewJob is to update the configuration value and submit the request to create the new job
  saveNewJob(submit: boolean) {
    //update the configuration
    this.jobDetail.name = this.name;
    this.jobDetail.description = this.desc;
    this.jobDetail.project_uuid = this.projectUUID;
    this.jobDetail.initiator_data.data_uuid = this.self.associated_data.split('+')[1];
    //if job type is modeling job
    if (this.newJobType == "modeling") {
      this.jobDetail.type = 1;
      if (this.dslRef && this.alRef) {
        this.jobDetail.conf_json = JSON.stringify(this.alRef.jsonObj);
        this.jobDetail.dsl_json = JSON.stringify(this.dslRef.jsonObj);
      } else {
        this.jobDetail.conf_json = this.algorithmConfig;
        this.jobDetail.dsl_json = this.dsl;
      }
      this.jobDetail.training_model_name = this.model_name;
      if (this.validationDataPercent === "" || this.validationDataPercent === "0") {
        this.validationDataPercent = "0";
        this.jobDetail.training_validation_enabled = false;
      } else {
        this.jobDetail.training_validation_enabled = true;
      }
      this.jobDetail.training_validation_percent = Number(this.validationDataPercent);
      this.jobDetail.initiator_data.label_name = this.self.label_column;
      if (!this.dropOrJson) {
        this.jobDetail.evaluate_component_name = "Evaluation_0"
        if (this.algorithm === 'al1') {
          this.jobDetail.training_algorithm_type = 1;
          this.jobDetail.algorithm_component_name = 'HomoLR_0'
        }
        if (this.algorithm === 'al2') {
          this.jobDetail.training_algorithm_type = 2;
          this.jobDetail.algorithm_component_name = 'HomoSecureboost_0'
        }
        if (this.algorithm === 'al3') {
          this.jobDetail.training_algorithm_type = 3;
          this.jobDetail.algorithm_component_name = 'HeteroLR_0'
        }
        if (this.algorithm === 'al4') {
          this.jobDetail.training_algorithm_type = 4;
          this.jobDetail.algorithm_component_name = 'HeteroSecureBoost_0'
        }
      } else {
        this.jobDetail.training_algorithm_type = this.JobAlgorithmType
        for (const data in this.svgData) {
          if (data.indexOf('HomoLR') !== -1 || data.indexOf('HomoSecureboost') !== -1 || data.indexOf('HeteroLR') !== -1 || data.indexOf('HeteroSecureBoost') !== -1) {
            this.jobDetail.algorithm_component_name = data
          } else if (data.indexOf('Evaluation') !== -1) {
            this.jobDetail.evaluate_component_name = data
          }
        }
      }
    } else if (this.newJobType == "predict") {
      this.jobDetail.type = 2;
      this.jobDetail.predicting_model_uuid = this.predictModel;
    } else {
      this.jobDetail.type = 3;
    }
    //if job type is predict job
    this.jobDetail.other_site_data = [];
    if (this.newJobType == "predict") {
      for (let party of this.newPredictParticipantList) {
        if (party.associated_data != "") {
          let selecteddata = {
            data_uuid: party.associated_data.split('+')[1],
            label_name: ""
          }
          this.jobDetail.other_site_data.push(selecteddata);
        }
      }
    } else {
      for (let party of this.participantList) {
        if (party.selected && party.associated_data != "") {
          let selecteddata = {
            data_uuid: party.associated_data.split('+')[1],
            label_name: ""
          }
          this.jobDetail.other_site_data.push(selecteddata);
        }
      }
    }
    this.jobDetail.training_component_list_to_deploy = []
    this.program.forEach(el => {
      if (el.value) {
        this.jobDetail.training_component_list_to_deploy.push(el.name)
      }
    })
    //if need to submit the request of creating a new job
    if (submit) {
      this.submitNewJob = true;
      this.submitNewJobFailed = false;
      this.submitNewJobFailed = this.checkSubmitValid(submit);
      if (!this.submitNewJobFailed) {
        this.projectservice.createJob(this.projectUUID, this.jobDetail)
          .subscribe(data => {
            this.msg.success('serverMessage.create200', 1000)
            const new_job_uuid = data.data.uuid;
            this.router.navigate(['project-management', 'project-detail', this.projectUUID, 'job', 'job-detail', new_job_uuid]);
          },
            err => {
              this.submitNewJobFailed = true;
              this.errorMessage = err.error.message;
            }
          );
      }
    }
  }

  //checkSubmitValid is to validate all configuration before submit the request of creating job
  checkSubmitValid(submit: boolean): boolean {
    //basic configuration
    if (this.jobDetail.name === '') {
      this.errorMessage = "Name can not be empty.";
      return true;
    }
    if (this.form.get('name')?.errors?.minlength || this.form.get('name')?.errors?.maxlength) {
      this.errorMessage = "Invalid name";
      return true;
    }
    if (this.jobDetail.initiator_data.data_uuid === '' || this.jobDetail.initiator_data.data_uuid === undefined) {
      this.errorMessage = "Please select initiator(self) data.";
      return true;
    }
    //if job type is modeling
    if (this.modeling) {
      if (this.jobDetail.training_model_name === '') {
        this.errorMessage = "Model name can not be empty.";
        return true;
      }
      if (!this.form.get('model_name')?.errors?.minlength || !this.form.get('model_name')?.errors?.maxlength) {
        this.errorMessage = this.form.get('model_name')?.errors?.emptyMessage || this.form.get('model_name')?.errors?.message;
      }
      if ((this.jobDetail.conf_json === '' || this.jobDetail.dsl_json === '') && submit) {
        this.errorMessage = "Please generate Algorithm Configuration and Workflow DSL.";
        return true;
      }
      if (this.algorithm === '') {
        this.errorMessage = "Please select a algorithm.";
        return true;
      }
    }
    //if job type is prediciton
    if (this.predict) {
      if (this.predictModel === '') {
        this.errorMessage = "Please select a model.";
        return true;
      }
      if (this.jobDetail.other_site_data.length != this.newPredictParticipantList.length) {
        this.errorMessage = "Please select participant data.";
        return true;
      }
    }
    return false;
  }

  // Generate logic diagram in json mode
  startDag(dsl: string, al: string) {
    this.dagJson = new DagJson(dsl, al, "#svg-canvas", "#component_info");
    this.dagJson.tooltip_css = "dagTip";
    this.dagJson.Generate();
    this.dagJson.Draw();
    this.program = this.dagJson.my
  }

  //generateConfig is to get or generate the dsl and algorithm configuration
  generateConfig(bool: boolean) {
    // Determine which mode is currently
    if (bool) {// drag
      if (!this.interactiveValidator()) return
      this.algorithm = '1'
      this.form.get('algorithm')?.setValue('1')
      this.setJsonOrDrag()
    } else {// json
      this.saveNewJob(false);
      this.submitGenerated = true;
      this.submitGeneratedFailed = this.checkSubmitValid(false);
      this.algorithmConfig = ''
      this.dsl = ''
      if (!this.submitGeneratedFailed) {
        this.projectservice.generateJobConfig(this.jobDetail)
          .subscribe(data => {
            this.algorithmConfig = data.data.conf_json;
            this.dsl = data.data.dsl_json;
            this.submitGeneratedFailed = false;
            this.program = []
            this.startDag(data.data.dsl_json, data.data.conf_json)
          },
            err => {
              this.submitGeneratedFailed = true;
              this.errorMessage = err.error.message;
            });
      }
    }
  }

  //showModelList is to get availble model list for prediction job
  showModelList() {
    this.projectservice.getModelList(this.projectUUID)
      .subscribe((data: any) => {
        this.modelList = data.data;
      },
        err => {
          this.isShowModelFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }
  //onSelectedModelChange is triggered when the model selection of predict job is changed
  onSelectedModelChange() {
    if (this.predictModel == '') {
      this.newPredictParticipantList = [];
      this.displayPredictParticipant = JSON.parse(JSON.stringify(this.newPredictParticipantList));
      return;
    }
    //get new available participant list for current model
    this.showPredictParticipant(this.predictModel);
  }

  //get available participant list for model
  showPredictParticipant(model_uuid: string) {
    this.newPredictParticipantList = [];
    if (model_uuid === '') return;
    this.projectservice.getPredictParticipant(model_uuid)
      .subscribe((data: any) => {
        this.predictParticipantList = data.data;
        //init the participant list for prediction job
        for (let party of this.predictParticipantList) {
          if (party.site_uuid != this.self.uuid) {
            const predictParticipant: PartyUser = {
              creation_time: "",
              description: "",
              name: party.site_name,
              party_id: party.site_party_id,
              status: 0,
              uuid: party.site_uuid,
              selected: true,
              associated_data: "",
              data_list: [],
              label_column: ""
            };
            this.getParticipantAssociatedDataList(predictParticipant, predictParticipant.selected, predictParticipant.uuid);
            this.newPredictParticipantList.push(predictParticipant);
          }
        }
        this.displayPredictParticipant = JSON.parse(JSON.stringify(this.newPredictParticipantList));
      },
        err => {
          this.isShowModelFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

}
