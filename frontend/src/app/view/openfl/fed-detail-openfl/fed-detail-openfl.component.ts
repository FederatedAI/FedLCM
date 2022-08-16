import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { OpenflService } from 'src/app/services/openfl/openfl.service';
import { LabelModel, EnvoyModel, DirectorInfoModel } from 'src/app/services/openfl/openfl-model-type'
import { ParticipantFATEStatus, ParticipantFATEType, constantGather } from 'src/utils/constant';
import { FormBuilder, FormGroup } from '@angular/forms';
import { ValidatorGroup } from 'src/utils/validators'
import { CreateTokenType } from 'src/app/services/openfl/openfl-model-type'
import * as moment from 'moment'
@Component({
  selector: 'app-fed-detail-openfl',
  templateUrl: './fed-detail-openfl.component.html',
  styleUrls: ['./fed-detail-openfl.component.scss', './fed-detail-openfl.component2.scss', './fed-detail-openfl.component3.scss']
})
export class FedDetailOpneFLComponent implements OnInit {

  newTokenForm!: FormGroup
  /* token related data */
  tokenList: any = []
  labelKey = ''
  labelValue = ''
  labelList: LabelModel[] = []
  // expired_at
  date!: any
  minDate = moment(Date.now()).format('YYYY-MM-DD')
  // 
  showLabelListTop = 0

  // token labels filter
  searchList: { key: string, value: string }[] = [{
    key: '',
    value: ''
  }]
  // Filters already added
  filterString = ''

  /* Operation flag */
  newTokenModal = false
  newTokenLoading = false
  openflConfigFlag = false
  createTokenFail = false
  showSearchFlag = false
  addLabelFlag = false
  tokenLabelEnterflag = false
  isShowOpenflDetailFailed: boolean = false;
  isPageLoading: boolean = true;
  isEnvoyFlag = true
  isShowParticipantListFailed: boolean = false;
  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  envoyConfigLoading = false
  // CodeMirror Instance
  code: any

  // The utility class is bound to this
  participantFATEstatus = ParticipantFATEType;
  participantFATEtype = ParticipantFATEStatus;
  // bind the utility function to this，
  // Display the corresponding string according to the value returned by the constantGather backend
  constantGather = constantGather;

  /* director and envoy  related data */
  director!: DirectorInfoModel;
  directorAccessInfoList: { [key: string]: any }[] = []
  envoylist: EnvoyModel[] = [];
  envoyAccessInfoList: { [key: string]: any }[] = []

  // store all envoys
  storageDataList: EnvoyModel[] = [];

  // openfl federation uuid
  uuid = String(this.route.snapshot.paramMap.get('id'));
  openflFederationDetail: any;
  errorMessage = "Service Error!"

  // deleted content
  deleteType: 'token' | 'director' | 'envoy' | 'federation' | 'multipleEnvoy' = 'token';
  deleteUUID: string = ''
  forceRemove = false
  // tabs flag
  get isEnvoy() {
    return this.isEnvoyFlag
  }
  set isEnvoy(value) {
    this.isEnvoyFlag = value
    this.filterString = ''
    this.showSearchOptions()
    this.showSearchFlag = false
  }
  get clientDisabled() {
    if (this.director && this.director.status === 1) {
      return false
    }
    return true
  }

  get createTokenDisabled() {
    return !this.newTokenForm.valid
  }

  get submitFilterDisbabled() {
    return this.searchList.every(el => {
      if (el.key === '' || el.value === '') {
        return false
      } else {
        return true
      }
    })
  }
  constructor(private openflService: OpenflService, private router: Router, private route: ActivatedRoute, private fb: FormBuilder) {
    this.newTokenForm = this.fb.group(
      ValidatorGroup([
        {
          name: 'tokenName',
          value: '',
          type: ['word'],
          max: 20,
          min: 2
        },
        {
          name: 'description',
          value: '',
          type: ['']
        },
        {
          name: 'expirationDate',
          value: '',
          type: ['require']
        },
        {
          name: 'limit',
          value: '',
          type: ['number']
        },

      ])
    )
  }

  ngOnInit(): void {
    this.isPageLoading = true;
    this.showOpenflDetail(this.uuid);
    this.showTokenList(this.uuid)
    this.showParticipantDetail()
  }
  // get Openfl Detail
  showOpenflDetail(uuid: string) {
    this.isShowOpenflDetailFailed = false;
    // get openfl detail
    this.openflService.getOpenflFederationDetail(uuid)
      .subscribe((data: any) => {
        this.openflFederationDetail = data.data;
        this.openflFederationDetail.fileList = []
        // Convert file object to array
        for (const key in data.data.shard_descriptor_config.python_files) {
          this.openflFederationDetail.fileList.push(key)
        }
        if (this.openflFederationDetail.use_customized_shard_descriptor) {
          this.openflToggleChange()
        }
      },
        err => {
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
          this.isShowOpenflDetailFailed = true;
        }
      );
  }

  // get director and envoy list
  showParticipantDetail() {
    this.openflService.getParticipantInfo(this.uuid).subscribe(
      data => {
        this.director = data.data.director
        this.envoylist = data.data.envoy
        this.storageDataList = data.data.envoy
        if (this.director) {
          for (const key in this.director.access_info) {
            const obj: any = {
              name: key,
            }
            const value = this.director.access_info[key]
            if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
              for (const key2 in value) {
                obj[key2] = value[key2]
              }
            }
            this.directorAccessInfoList.push(obj)
          }
        }
        if (this.envoylist) {
          this.envoylist.forEach((envoy: any) => {
            const labels = []
            for (const key in envoy.labels) {
              const obj: any = {
                key: key,
                value: envoy.labels[key]
              }
              labels.push(obj)
            }
            envoy.labels = labels
            const access_info: any[] = []
            if (envoy.access_info && this.hasAccessInfo(envoy.access_info)) {
              for (const key in envoy.access_info) {
                const obj = {
                  name: key,
                  value: envoy.access_info[key]
                }
                access_info.push(obj)
              }
              envoy.access_info = access_info
            }
          })
        }
      },
      err => {
        this.errorMessage = err.error.message;
        this.isPageLoading = false;
        this.isShowOpenflDetailFailed = true;
      }
    )
  }
  // get token List
  showTokenList(uuid: string) {
    // this.isPageLoading = true;
    this.isShowParticipantListFailed = false;
    this.openflService.getTokenList(uuid)
      .subscribe((data: any) => {
        this.tokenList = []
        data.data?.forEach((el: any) => {
          const labels = JSON.parse(JSON.stringify(el.labels))
          el.labels = []
          for (const key in labels) {
            el.labels.push({
              key: key,
              value: labels[key]
            })
          }
          this.tokenList.push(el)
        });
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
          this.isShowParticipantListFailed = true;
        }
      );
  }

  // add token label
  addLabel() {
    this.labelList.push({
      key: this.labelKey,
      value: this.labelValue,
      no: Date.now()
    })
    this.labelValue = ''
    this.labelKey = ''
    this.addLabelFlag = false
  }
  // close new label
  cancelLabel() {
    this.labelValue = ''
    this.labelKey = ''
    this.addLabelFlag = false
  }
  delLabel(index: number) {
    this.labelList.splice(index, 1)
  }
  // envoy label mouserEnter 
  mouseEnter(enovy: EnvoyModel) {
    if (enovy.labels.length < 3) {
      this.showLabelListTop = 0
    } else {
      this.showLabelListTop = Math.ceil((enovy.labels.length - 3)) * 20
    }
    enovy.showLabelListFlag = true
  }
  // submit token
  createToken() {
    const tokenInfo: CreateTokenType = {
      description: this.newTokenForm.get('description')?.value,
      expired_at: this.date,
      labels: {},
      limit: this.newTokenForm.get('limit')?.value * 1,
      name: this.newTokenForm.get('tokenName')?.value
    }
    this.labelList.forEach(el => [
      tokenInfo.labels[el.key] = el.value
    ])
    if (this.labelKey && this.labelValue) {
      tokenInfo.labels[this.labelKey] = this.labelValue
    }
    this.openflService.createTokenInfo(this.uuid, tokenInfo).subscribe(
      data => {
        this.showTokenList(this.uuid)
        this.resetToken()
        this.newTokenModal = false
      },
      err => {
        this.newTokenModal = true
        this.createTokenFail = true
        this.errorMessage = err.error.message

      }
    )
  }
  // reset token modal
  resetToken() {
    this.newTokenForm.reset()
    this.date = new Date()
    this.labelList = []
    this.createTokenFail = false
  }

  // Judge whether the object is empty
  hasAccessInfo(object: any): boolean {
    return JSON.stringify(object) !== '{}'
  }

  // create customize
  createCustomize(yamlHTML: any) {
    if (!this.code) {
      this.code = window.CodeMirror.fromTextArea(yamlHTML, {
        value: '',
        mode: 'yaml',
        lineNumbers: true,
        indentUnit: 1,
        lineWrapping: true,
        tabSize: 2,
        readOnly: true
      })
    }
    if (this.code) {
      this.code.setValue(this.openflFederationDetail.shard_descriptor_config.envoy_config_yaml)
      this.envoyConfigLoading = false
    }
  }

  // customize change
  openflToggleChange() {
    this.envoyConfigLoading = true
    setTimeout(() => {
      const yamlHTML = document.getElementById('yaml') as any
      if (!yamlHTML) {
        try {
          const timer = setInterval(() => {
            const yamlHTML = document.getElementById('yaml') as any
            if (yamlHTML) {
              this.createCustomize(yamlHTML)
              clearInterval(timer)
            }
          }, 100)
        } catch (error) {
        }
      } else {
        try {
          this.createCustomize(yamlHTML)
        } catch (error) {
        }
      }
    });
  }

  multipleDeletion = false;
  openDeleteConfrimModal(type: 'token' | 'director' | 'envoy' | 'federation' | 'multipleEnvoy', item_uuid: string) {
    this.deleteType = type;
    this.deleteUUID = item_uuid
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
    this.forceRemove = false;
    this.multipleDeletion = type === 'multipleEnvoy' ? true : false;
    if (this.seletedEnvoyList.length > 1) this.multipleDeletion = true;
  }

  // delete submit
  delete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    if (this.deleteType === 'federation') {
      this.openflService.deleteOpenflFederation(this.uuid).subscribe(
        data => {
          this.isDeleteSubmit = false;
          this.isDeleteFailed = false;
          this.openDeleteModal = false
          this.router.navigate(['/federation'])
        },
        error => {
          this.openDeleteModal = true
          this.isDeleteSubmit = true;
          this.isDeleteFailed = true;
          this.errorMessage = error.error.message
        }
      )
    } else if (this.deleteType === 'token') {
      this.openflService.deleteTokenInfo(this.uuid, this.deleteUUID).subscribe(
        data => {
          this.isDeleteSubmit = false;
          this.isDeleteFailed = false;
          this.openDeleteModal = false
          this.showTokenList(this.uuid)
        },
        error => {
          this.openDeleteModal = true
          this.isDeleteSubmit = true;
          this.isDeleteFailed = true;
          this.errorMessage = error.error.message
        }
      )
    } else if (this.deleteType === 'director') {
      this.openflService.deleteDirector(this.uuid, this.deleteUUID, this.forceRemove).subscribe(
        data => {
          this.isDeleteSubmit = false;
          this.isDeleteFailed = false;
          this.openDeleteModal = false
          this.showParticipantDetail()
        },
        error => {
          this.openDeleteModal = true
          this.isDeleteSubmit = true;
          this.isDeleteFailed = true;
          this.errorMessage = error.error.message
        }
      )
    } else if (this.deleteType === 'envoy' && !this.multipleDeletion) {
      this.openflService.deleteEnvoy(this.uuid, this.deleteUUID, this.forceRemove).subscribe(
        data => {
          this.isDeleteSubmit = false;
          this.isDeleteFailed = false;
          this.openDeleteModal = false
          this.showParticipantDetail()
        },
        error => {
          this.openDeleteModal = true
          this.isDeleteSubmit = true;
          this.isDeleteFailed = true;
          this.errorMessage = error.message
        }
      )
    } else if (this.deleteType === 'multipleEnvoy' || this.multipleDeletion) {
      for (let envoy of this.seletedEnvoys) {
        envoy.deleteFailed = false;
        envoy.deleteSuccess = false;
        this.isDeleteFailed = false;
        envoy.deleteSubmit = true;
        this.isDeleteSubmit = true;
        this.openflService.deleteEnvoy(this.uuid, envoy.uuid, this.forceRemove).subscribe(
          data => {
            envoy.deleteFailed = false;
            envoy.deleteSuccess = true;
            this.isDeleteFailed = false;
          },
          error => {
            this.openDeleteModal = true
            envoy.deleteFailed = true;
            envoy.deleteSuccess = false;
            this.isDeleteFailed = true;
            envoy.errorMessage = error.message
          }
        )
      }
    }
  }

  // Select all logos
  get allSelect() {
    if (this.envoylist && this.envoylist.length > 0) {
      return this.envoylist.every(el => el.selected === true)
    } else {
      return false
    }
  }
  set allSelect(val) {
    this.envoylist.forEach(el => el.selected = val)
  }

  seletedEnvoyList: EnvoyModel[] = []
  get seletedEnvoys() {
    this.seletedEnvoyList = []
    if (this.envoylist && this.envoylist.length > 0) {
      for (const envoy of this.envoylist) {
        if (envoy.selected && envoy.status != 2) {
          this.seletedEnvoyList.push(envoy)
        }
      }
    }
    return this.seletedEnvoyList
  }

  isDeleteEvonyAllSuccess = false;
  get deleteEvonyAllSuccess() {
    for (const envoy of this.seletedEnvoys) {
      if (!envoy.deleteSuccess) {
        this.isDeleteEvonyAllSuccess = false
        return this.isDeleteEvonyAllSuccess
      }
    }
    this.isDeleteEvonyAllSuccess = true
    return this.isDeleteEvonyAllSuccess
  }

  //refresh is for refresh button
  refresh() {
    this.showOpenflDetail(this.uuid)
    this.showTokenList(this.uuid)
    this.showParticipantDetail()
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
  // Jump envoy details
  toDetail(type: string, detailId: string, info: any) {
    this.route.params.subscribe(
      value => {
        this.router.navigateByUrl(`federation/openfl/${value.id}/${type}/detail/${detailId}`)
      }
    )
  }
  // show envoy filter UI
  showSearchOptions() {
    this.showSearchFlag = !this.showSearchFlag
    this.searchList = [{
      key: '',
      value: ''
    }]
  }
  // reset envoy filter
  recycleSearchOptions() {
    this.showSearchFlag = false
    this.filterString = ''
    this.envoylist = this.storageDataList
    this.searchList = [{
      key: '',
      value: ''
    }]
  }
  // add envoy search label
  addSearchOption() {
    this.searchList.push({
      key: '',
      value: ''
    })
  }
  // delete envoy search label
  delSearchOption(index: number) {
    this.searchList.splice(index, 1)
  }
  // confirm filter
  submitFilter() {
    this.filterString = ''
    const keys: string[] = []
    const values: string[] = []
    const filterEnvoyKeyList: EnvoyModel[] = []
    const filterEnvoyList: EnvoyModel[] = []
    this.searchList.forEach(el => {
      keys.push(el.key)
      values.push(el.value)
      this.filterString += `${el.key}:${el.value},`
    })
    this.storageDataList.filter(el => {
      el.labels.forEach(label => {
        if (keys.find(key => key === label.key)) {
          if (!filterEnvoyKeyList.find(envoy => envoy.uuid === el.uuid)) {
            filterEnvoyKeyList.push(el)
          }
        }
      })
    })
    filterEnvoyKeyList.forEach(el => {
      el.labels.forEach(label => {
        if (values.find(value => value === label.value)) {
          if (!filterEnvoyList.find(envoy => envoy.uuid === el.uuid)) {
            filterEnvoyList.push(el)
          }
        }
      })
    })
    this.envoylist = filterEnvoyList
    this.showSearchFlag = false
  }
}
