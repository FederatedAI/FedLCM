import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router'
import { FedService } from 'src/app/services/federation-fate/fed.service'
import { OpenflService } from 'src/app/services/openfl/openfl.service';
import { constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-envoy-detail',
  templateUrl: './envoy-detail.component.html',
  styleUrls: ['./envoy-detail.component.scss']
})
export class EnvoyDetailComponent implements OnInit {
  isShowDetailFailed = false
  isPageLoading = true
  errorMessage = ''
  uuid = ''
  envoy_uuid = ''
  deleteType = 'cluster'
  forceRemove = false
  constantGather = constantGather
  envoyDetail: any = {}
  openDeleteModal = false
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  labelsList: { [key: string]: any }[] = []
  constructor(private route: ActivatedRoute, private router: Router, private openflService: OpenflService) { }
  ngOnInit(): void {
    this.getEnvoyDetail()
  }
  code: any
  overview = true
  get isOverview() {
    return this.overview
  }
  set isOverview(value) {
    if (value) {
      const yamlHTML = document.getElementById('yaml') as any
      this.code = window.CodeMirror.fromTextArea(yamlHTML, {
        value: '',
        mode: 'yaml',
        lineNumbers: true,
        indentUnit: 1,
        lineWrapping: true,
        tabSize: 2,
        readOnly: true
      })
      if (this.envoyDetail.deployment_yaml) {
        this.code.setValue(this.envoyDetail.deployment_yaml)
      }
    } else {
      this.code = null
    }
    this.overview = value
  }
  getEnvoyDetail() {
    this.openDeleteModal = false
    this.isDeleteSubmit = false;
    this.isDeleteFailed = false;
    this.isShowDetailFailed = false
    this.isPageLoading = true
    this.errorMessage = ''
    this.uuid = ''
    this.envoy_uuid = ''
    this.labelsList = []
    this.route.params.subscribe(
      value => {
        this.uuid = value.id
        this.envoy_uuid = value.envoy_uuid
        //test
        this.isPageLoading = false
        if (this.uuid && this.envoy_uuid) {
          this.openflService.getEnvoyInfo(this.uuid, this.envoy_uuid).subscribe(
            data => {
              this.envoyDetail = data.data
              for (const key in data.data.labels) {
                const obj: any = {
                  key: key,
                  value: data.data.labels[key]
                }
                this.labelsList.push(obj)
              }
              this.isPageLoading = false

            },
            err => {
              this.isPageLoading = false
              this.errorMessage = err.error.message;
              this.isShowDetailFailed = false
            }
          )
        }
      }
    )
  }

  //refresh is for refresh button
  refresh() {
    this.getEnvoyDetail()
  }

  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal() {
    this.openDeleteModal = true
    this.forceRemove = false;
    this.isDeleteSubmit = false;
    this.isDeleteFailed = false;
  }

  confirmDelete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.openflService.deleteEnvoy(this.uuid, this.envoy_uuid, this.forceRemove)
      .subscribe(() => {
        this.router.navigate(['/federation/openfl', this.uuid]);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });

  }
}
