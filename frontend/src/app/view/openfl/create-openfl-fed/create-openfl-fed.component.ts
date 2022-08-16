import { Component, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { ValidatorGroup } from 'src/utils/validators'
import { OpenflPsotModel } from 'src/app/services/openfl/openfl-model-type'
import { OpenflService } from 'src/app/services/openfl/openfl.service'
import { Observable, throwError } from 'rxjs'

@Component({
  selector: 'app-create-openfl',
  templateUrl: './create-openfl-fed.component.html',
  styleUrls: ['./create-openfl-fed.component.scss']
})
export class CreateOpenflComponent implements OnInit {

  constructor(
    private fb: FormBuilder,
    private openflService: OpenflService
  ) { }

  ngOnInit(): void {
  }
  openflForm = this.fb.group(
    ValidatorGroup([
      {
        name: 'name',
        value: '',
        type: ['word'],
        max: 20,
        min: 2
      },
      {
        name: 'description',
        type: ['']
      },
      {
        name: 'domain',
        value: '',
        type: ['fqdn']
      },
      {
        name: 'customize',
        value: false,
        type: ['']
      },
      {
        name: 'target',
        value: '',
        type: ['number-list']
      },
      {
        name: 'sample',
        value: '',
        type: ['number-list']
      },
      {
        name: 'envoyYaml',
        value: '',
        type: ['require']
      }
    ])
  )
  code: any
  fileStatus = ''
  requirementStatus = ''
  requirements: any = {}
  uploadFileList: any[] = []


  openflToggleChange() {
    setTimeout(() => {
      try {
        this.setCodeMirror()
      } catch (error) {

      }
    })
  }
  // Create rich text
  setCodeMirror() {
    const yamlHTML = document.getElementById('yaml') as any
    this.code = window.CodeMirror.fromTextArea(yamlHTML, {
      value: '',
      mode: 'yaml',
      lineNumbers: true,
      indentUnit: 1,
      lineWrapping: true,
      tabSize: 2,
    })
    // Listen to the rich text input and save it to the form
    this.code.on('change', (cm: any) => {
      this.code.save()
      this.openflForm.get('envoyYaml')?.setValue(this.code.getValue())
    })
  }
  // Form validation when customize is false
  customizeFalseValidate() {
    const validList = ['name', 'domain']
    const result = validList.every(el => {
      const obj = this.openflForm.get(el)
      return obj && obj.valid === true
    })
    return result
  }

  // Form validation when customize is true
  customizeTrueValidate() {
    return this.openflForm.valid && this.fileStatus === 'success' && this.requirementStatus !== 'error'
  }

  uploadFileChange(e: any, fileType: string) {
    if (fileType === 'py') {
      if (e.target.files.length === 0) {
        this.fileStatus = 'enabled'
        this.uploadFileList = []
      } else {
        if (!this.isPythonFile(e.target.files)) {
          this.fileStatus = 'error'
        } else {
          for (let i = 0; i < e.target.files.length; i++) {
            // Create a FileReader
            const reader = new FileReader();
            // Read file
            reader.readAsText(e.target.files[i], "UTF-8");
            reader.onload = (evt: any) => {
              // Get file content
              const fileString = evt.target.result;
              this.uploadFileList.push({
                name: e.target.files[i].name,
                content: fileString
              })
            }
          }
          this.fileStatus = 'success'
        }
      }
    } else {
      this.requirementStatus = 'enabled'
      if (e.target.files.length !== 0) {
        if (this.isRequirementsFile(e.target.files)) {
          const reader = new FileReader();
          reader.readAsText(e.target.files[0], "UTF-8");
          reader.onload = (evt: any) => {
            // Get file content
            const fileString = evt.target.result;
            this.requirements = {
              name: 'requirements.txt',
              content: fileString
            }
          }
          this.requirementStatus = 'success'
        } else {
          this.requirementStatus = 'error'
        }
      } else {
        this.requirements = {}
      }
    }
  }

  // Determine whether it is a python file
  isPythonFile(files: any) {
    let result = true
    for (let i = 0; i < files.length; i++) {
      const suffix = files[i].name.slice(-3)
      if (suffix !== '.py') {
        result = false
        break
      }
    }
    return result
  }

  // Determine whether it is a requirements.txt
  isRequirementsFile(files: any): boolean {
    for (let i = 0; i < files.length; i++) {
      if (files[i].name !== 'requirements.txt') {
        return false
      }
    }
    return true
  }

  createNewOpenfl(): Observable<any> {
    const customize = this.openflForm.get('customize')?.value
    const openflInfo: OpenflPsotModel = {
      name: this.openflForm.get('name')?.value,
      description: this.openflForm.get('description')?.value,
      domain: this.openflForm.get('domain')?.value,
      shard_descriptor_config: {
        envoy_config_yaml: '',
        python_files: {},
        sample_shape: [],
        target_shape: []
      },
      use_customized_shard_descriptor: true
    }
    if (customize) {
      if (this.customizeTrueValidate()) {
        openflInfo.shard_descriptor_config.envoy_config_yaml = this.openflForm.get('envoyYaml')?.value
        openflInfo.shard_descriptor_config.sample_shape = this.openflForm.get('sample')?.value.split(',')
        openflInfo.shard_descriptor_config.target_shape = this.openflForm.get('target')?.value.split(',')
        this.uploadFileList.forEach(el => {
          openflInfo.shard_descriptor_config.python_files[el.name] = el.content
        })
        if (this.requirements.name) {
          openflInfo.shard_descriptor_config.python_files[this.requirements.name] = this.requirements.content
        }
        return this.openflService.createOpenflFederation(openflInfo)
      } else {
        return throwError('Verification failed')
      }
    } else {
      if (this.customizeFalseValidate()) {
        openflInfo.use_customized_shard_descriptor = false
        return this.openflService.createOpenflFederation(openflInfo)
      } else {
        return throwError('Verification failed')
      }
    }

  }
}
