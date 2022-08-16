import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router'
import { OpenflService } from 'src/app/services/openfl/openfl.service';
import { constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-director-detail',
  templateUrl: './director-detail.component.html',
  styleUrls: ['./director-detail.component.scss']
})
export class DirectorDetailComponent implements OnInit {
  isShowDetailFailed = false
  isPageLoading = true
  errorMessage = "Service Error!"
  uuid = ''
  director_uuid = ''
  constantGather = constantGather
  directorDetail: any = {}
  openDeleteModal = false
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  accessInfoList: { [key: string]: any }[] = []
  constructor(private route: ActivatedRoute, private router: Router, private openflService: OpenflService) { }
  ngOnInit(): void {
    this.getDirectorDetail()
  }
  code: any
  overview = true
  get isOverview() {
    return this.overview
  }
  set isOverview(value) {
    if (value) {
      const yamlHTML = document.getElementById('yaml') as any
      if (this.directorDetail.deployment_yaml) {
        this.createCustomize(yamlHTML, this.directorDetail.deployment_yaml)
      }
    } else {
      this.code = null
    }
    this.overview = value
  }

  getDirectorDetail() {
    this.openDeleteModal = false
    this.isDeleteSubmit = false;
    this.isDeleteFailed = false;
    this.isShowDetailFailed = false
    this.isPageLoading = true
    this.errorMessage = ''
    this.uuid = ''
    this.director_uuid = ''
    this.route.params.subscribe(
      value => {
        this.uuid = value.id
        this.director_uuid = value.director_uuid
        if (this.uuid && this.director_uuid) {
          this.openflService.getDirectorInfo(value.id, value.director_uuid).subscribe(
            data => {
              this.directorDetail = data.data
              const value = data.data.deployment_yaml
              setTimeout(() => {
                const yamlHTML = document.getElementById('yaml') as any
                if (!yamlHTML) {
                  try {
                    const timer = setInterval(() => {
                      const yamlHTML = document.getElementById('yaml') as any
                      if (yamlHTML) {
                        this.createCustomize(yamlHTML, value)
                        clearInterval(timer)
                      }
                    }, 100)
                  } catch (error) {
                  }
                } else {
                  try {
                    this.createCustomize(yamlHTML, value)
                  } catch (error) {
                  }
                }
              })
              for (const key in data.data.access_info) {
                const obj: any = {
                  name: key,
                }
                const value = data.data.access_info[key]
                if (Object.prototype.toString.call(value).slice(8, -1) === 'Object') {
                  for (const key2 in value) {
                    obj[key2] = value[key2]
                  }
                }
                this.accessInfoList.push(obj)
              }
              this.isPageLoading = false
            },
            err => {
              this.isPageLoading = false
              this.errorMessage = err.error.message;
              this.isShowDetailFailed = true
            }
          )
        }
      }
    )

  }

  // create customize
  createCustomize(yamlHTML: any, data: string) {
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
    setTimeout(() => {
      this.code.setValue(data)
      this.isPageLoading = false
    })
  }

  //refresh is for refresh button
  refresh() {
    this.accessInfoList = []
    this.getDirectorDetail()
  }

  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal() {
    this.openDeleteModal = true
    this.forceRemove = false
  }
  get accessInfo() {
    return JSON.stringify(this.directorDetail.access_info) === '{}'
  }
  deleteType = 'director'
  forceRemove = false
  confirmDelete() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.openflService.deleteDirector(this.uuid, this.director_uuid, this.forceRemove)
      .subscribe(() => {
        this.router.navigate(['/federation/openfl', this.uuid]);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });

  }
  toLink(item: any) {
    window.open(
      `http://${item.host}:${item.port}`, '_blank'
    )
  }
}
