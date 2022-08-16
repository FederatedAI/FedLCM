import { Component, OnInit } from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { EndpointService } from 'src/app/services/common/endpoint.service';
import { OpenflService } from 'src/app/services/openfl/openfl.service'
import { ENDPOINTSTATUS, CHARTTYPE, constantGather } from 'src/utils/constant';
import { ChartService } from 'src/app/services/common/chart.service';
import { ValidatorGroup } from 'src/utils/validators'
import { InfraService } from 'src/app/services/common/infra.service';
import { DirectorModel } from 'src/app/services/openfl/openfl-model-type';
import { EndpointType } from 'src/app/view/endpoint/endpoint-model'

@Component({
  selector: 'app-exchange-new',
  templateUrl: './director-new.component.html',
  styleUrls: ['./director-new.component.scss']
})
export class DirectorNewComponent implements OnInit {
  form: FormGroup;
  // The utility class is bound to this
  endpointStatus = ENDPOINTSTATUS;
  chartType = CHARTTYPE;
  // bind the utility function to this，
  // Display the corresponding string according to the value returned by the constantGather backend
  constantGather = constantGather

  // Display error related properties
  errorMessage: any;
  isShowEndpointFailed: boolean = false;
  isShowInfraDetailFailed: boolean = false;
  isShowChartFailed: boolean = false;
  isGenerateFailed = false;
  isCreatedFailed = false;

  // Action related properties
  isPageLoading: boolean = true;
  noEndpoint = false;
  openInfraModal: boolean = false;
  use_cert: boolean = false;
  skip_cert: boolean = false;
  endpointSelectOk: boolean = false;
  isGenerateSubmit = false;
  isCreatedSubmit = false;

  // Id
  openfl_uuid = String(this.route.snapshot.paramMap.get('id'));
  infraUUID: any = ""


  // The currently selected endpoint object
  endpoint!: EndpointType | null
  endpointlist: any = [];
  chartlist: any = [];

  // Intermediate amount
  infraConfigDetail: any;
  hasRegistry = false;
  hasRegistrySecret = false;
  registrySecretConfig: any;
  // CodeMirror instance
  codeMirror: any

  get selectedEndpoint() {
    return this.endpoint
  }
  set selectedEndpoint(value) {
    this.endpoint = value
    if (value) {
      this.onSelectEndpoint()
    }
  }
  get useRegistrySecret() {
    return this.form.controls['registry'].get('useRegistrySecret')?.value;
  }
  get useRegistry() {
    return this.form.controls['registry'].get('useRegistry')?.value;
  }
  get registryConfig() {
    return this.form.controls['registry'].get('registryConfig')?.value;
  }

  get server_url_suggestion() {
    var url_suggestion = "";
    var header = "https://"
    if (this.registryConfig === "" || this.registryConfig === null || this.registryConfig === "-") {
      url_suggestion = header + "index.docker.io/v1/"
    } else {
      var url = this.registryConfig.split('/')[0]
      if (this.validURL(url)) {
        url_suggestion = header + url
      } else {
        url_suggestion = header + "index.docker.io/v1/"
      }
    }
    return url_suggestion
  }

  get valid_server_url() {
    return this.form.get('registry.server_url')?.valid && this.form.controls['registry'].get('server_url')?.value?.trim() != ''
  }
  
  //cert_disabled is to validate the configuration of certificate section and disabled the 'Next' button if needed
  get cert_disabled() {
    return this.use_cert && (this.form.controls['certificate'].get('jupyter_client_cert_info_mode')?.value === 1
      || this.form.controls['certificate'].get('director_server_cert_info_mode')?.value === 1)
  }

  get service_type_disabled() {
    return !this.form.controls['serviceType'].get('serviceType')?.valid
  }

  get registry_disabled() {
    var registry_secret_valid: any = true;
    if (this.useRegistrySecret) {
      registry_secret_valid = (this.form.controls['registry'].get('username')?.value?.trim() != '') && (this.form.controls['registry'].get('password')?.value?.trim() != '') && this.form.get('registry.server_url')?.valid && (this.form.controls['registry'].get('server_url')?.value?.trim() != '')
    } else {
      registry_secret_valid = true
    }
    var registry_valid = true;
    if (this.useRegistry) {
      registry_valid = this.form.controls['registry'].get('registryConfig')?.value?.trim() != ''
    } else {
      registry_valid = true
    }
    return !(registry_secret_valid && registry_valid)
  }

  //submitDisable returns if disabled the submit button of create a new director
  get submitDisable() {
    if (!this.form.valid || this.cert_disabled || this.service_type_disabled || !this.isGenerateSubmit || (this.isCreatedSubmit && !this.isCreatedFailed)) {
      return true
    } else {
      return false
    }
  }

  constructor(private formBuilder: FormBuilder,
    private router: Router,
    private route: ActivatedRoute,
    private endpointService: EndpointService,
    private chartservice: ChartService,
    private infraservice: InfraService,
    private openflService: OpenflService
  ) {
    this.form = this.formBuilder.group({
      info: this.formBuilder.group({
        name: [''],
        description: [''],
      }),
      endpoint: this.formBuilder.group({
        endpoint_uuid: ['']
      }),
      chart: this.formBuilder.group({
        chart_uuid: [''],
      }),
      namespace: this.formBuilder.group({
        namespace: ['openfl-director'],
      }),
      certificate: this.formBuilder.group({
        cert: ['skip'],
        director_server_cert_info_mode: [1],
        director_server_cert_info_uuid: [''],
        director_server_cert_info_name: [''],
        jupyter_client_cert_info_uuid: [''],
        jupyter_client_cert_info_mode: [1],
        jupyter_client_cert_info_name: ['']
      }),
      jupyter: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'password',
            type: ['require'],
            value: ''
          }
        ])
      ),
      serviceType: this.formBuilder.group({
        serviceType: [null],
      }),
      registry: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'useRegistry',
            type: [''],
            value: false
          },
          {
            name: 'useRegistrySecret',
            type: [''],
            value: false
          },
          {
            name: 'registryConfig',
            type: [''],
            value: ''
          },
          {
            name: 'server_url',
            type: ['internet'],
            value: ''
          },
          {
            name: 'username',
            type: [''],
            value: ''
          },
          {
            name: 'password',
            type: [''],
            value: ''
          },
        ])
      ),
      psp: this.formBuilder.group({
        enablePSP: [false],
      }),
      yaml: this.formBuilder.group({
        yaml: [''],
      })
    });
  }
  ngOnInit(): void {
    this.showEndpointList();
    this.showChartList();
  }
  // Get Endpoint List
  showEndpointList() {
    this.isShowEndpointFailed = false;
    this.isPageLoading = true;
    this.endpointService.getEndpointList()
      .subscribe((data: any) => {
        this.endpointlist = data.data.filter((el: any) => { if (el.status === 2) return el });
        if (this.endpointlist.length === 0) {
          this.noEndpoint = true;
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowEndpointFailed = true;
          this.isPageLoading = false;
        });
  }
  // Get Chart List
  showChartList() {
    this.chartlist = [];
    this.isPageLoading = true;
    this.isShowChartFailed = false;
    this.chartservice.getChartList()
      .subscribe((data: any) => {
        if (data.data) {
          for (let chart of data.data) {
            if (chart.type === 3) {
              this.chartlist.push(chart);
            }
          }
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
  }
  //setRadioDisplay is trigger when the selection of certificate's mode is changed
  setRadioDisplay(val: any) {
    if (val == "use") {
      this.use_cert = true;
      this.skip_cert = false;
    }
    if (val == "skip") {
      this.use_cert = false;
      this.skip_cert = true;
      this.form.get('certificate')?.get('director_server_cert_info_mode')?.setValue(1);
      this.form.get('certificate')?.get('jupyter_client_cert_info_mode')?.setValue(1);
    }
  }

  // initYAMLEditorByEvent is triggered when clicking the yaml textarea(if there is no highlight) to reinitialize the YAML editor
  initYAMLEditorByEvent(event: any) {
    if (event && event.target && event.target.value) {
      this.isGenerateSubmit = true;
      this.initCodeMirror(event.target.value)
    }
  }

  // initCodeMirror is to initialize Code Mirror YAML editor
  initCodeMirror(yamlContent: any) {
    const yamlHTML = document.getElementById('yaml') as any
    this.codeMirror = window.CodeMirror.fromTextArea(yamlHTML, {
      value: yamlContent,
      mode: 'yaml',
      lineNumbers: true,
      indentUnit: 1,
      lineWrapping: true,
      tabSize: 2,
    })
    this.codeMirror.on('change', (cm: any) => {
      this.codeMirror.save()
    })
    this.hasYAMLTextAreaDOM = true
  }

  hasYAMLTextAreaDOM = false
  // onClickNext is triggered when clicking the 'Next' button before the step 'YAML'
  onClickNext() {
    this.hasYAMLTextAreaDOM = false;
    // when the form is reset by the clarity stepper component and the step 'YAML' is collapsed, the DOM of step 'YAML' will be cleared. 
    // So we need to check if we need to initialized code mirror window in the next step (step 'YAML)
    if (document.getElementById('yaml') !== null) {
      this.hasYAMLTextAreaDOM = true;
    }
  }

  get editorModeHelperMessage() {
    return this.codeMirror && !this.hasYAMLTextAreaDOM && this.form.controls['yaml'].get('yaml')?.value
  }
  
  //reset form when selection change
  onSelectEndpoint() {
    if (this.selectedEndpoint) {
      this.endpointSelectOk = true;
      this.form.get('endpoint')?.get('endpoint_uuid')?.setValue(this.selectedEndpoint.uuid);
      this.infraUUID = this.endpoint?.infra_provider_uuid
      this.showInfraDetail(this.infraUUID)
    }
  }
  // Get Infran Detail
  showInfraDetail(uuid: string) {
    this.isShowInfraDetailFailed = false;
    this.infraservice.getInfraDetail(uuid)
      .subscribe((data: any) => {
        this.infraConfigDetail = data.data.kubernetes_provider_info;
        this.hasRegistry = this.infraConfigDetail.registry_config_fate.use_registry
        this.hasRegistrySecret = this.infraConfigDetail.registry_config_fate.use_registry_secret
        this.registrySecretConfig = this.infraConfigDetail.registry_config_fate.registry_secret_config
        this.form.get('registry')?.get('useRegistrySecret')?.setValue(this.hasRegistrySecret);
        this.form.get('registry')?.get('useRegistry')?.setValue(this.hasRegistry);

        this.change_use_registry()
        this.change_registry_secret()
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowInfraDetailFailed = true;
        }
      );
  }
  // Use Registry change
  onChange_use_registry(val: any) {
    this.change_use_registry()
    this.selectChange(val)
  }

  change_use_registry() {
    if (this.useRegistry) {
      this.form.get('registry')?.get('registryConfig')?.setValue(this.infraConfigDetail.registry_config_fate.registry);
    } else {
      this.form.get('registry')?.get('registryConfig')?.setValue("-");
    }
  }
  // Use Registry Secret change
  onChange_use_registry_secret(val: any) {
    this.change_registry_secret()
    this.selectChange(val)
  }

  change_registry_secret() {
    if (this.useRegistrySecret) {
      this.form.get('registry')?.get('server_url')?.setValue(this.registrySecretConfig.server_url);
      this.form.get('registry')?.get('username')?.setValue(this.registrySecretConfig.username);
      this.form.get('registry')?.get('password')?.setValue(this.registrySecretConfig.password);
    }
    else {
      this.form.get('registry')?.get('server_url')?.setValue("https://x");
      this.form.get('registry')?.get('username')?.setValue("-");
      this.form.get('registry')?.get('password')?.setValue("-");
    }
  }

  formResetChild(child: string) {
    this.form.controls[child].reset();
    this.form.controls[child].reset();
    if (child === 'yaml' && this.codeMirror) {
      this.codeMirror.setValue('')
    }
  }

  validURL(str: string) {
    var pattern = new RegExp(
      '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|' + // domain name
      '((\\d{1,3}\\.){3}\\d{1,3}))' + // OR ip (v4) address
      '(\\?[;&a-z\\d%_.~+=-]*)?' + // query string
      '(\\#[-a-z\\d_]*)?$', 'i'); // fragment locator
    return !!pattern.test(str);
  }

  // After the content of the previous step has changed
  selectChange(val: any) {
    this.formResetChild('yaml');
    this.isGenerateSubmit = false;
    this.isGenerateFailed = false;
  }

  get enablePSP() {
    return this.form.controls['psp'].get('enablePSP')?.value;
  }

  //generateYaml() is to get the exchange initial yaml based on the configuration provided by user
  generateYaml() {
    const yamlHTML = document.getElementById('yaml') as any
    this.isGenerateSubmit = true;
    var chart_id = this.form.controls['chart'].get('chart_uuid')?.value;
    var namespace = this.form.controls['namespace'].get('namespace')?.value;
    var name = this.form.controls['info'].get('name')?.value;
    var service_type = Number(this.form.controls['serviceType'].get('serviceType')?.value);
    var password = this.form.controls['jupyter'].get('password')?.value;
    if (namespace === '') namespace = 'openfl-director';
    this.openflService.getDirectorYaml(this.openfl_uuid, password, chart_id, namespace, name, service_type, this.registryConfig, this.useRegistry, this.useRegistrySecret, this.enablePSP).subscribe(
      data => {
        this.form.get('yaml')?.get('yaml')?.setValue(data.data);
        // if code mirror object and yaml DOM are existing, just set value to the code mirror object
        // else initialize code mirror object and yaml editor window
        if (this.codeMirror && this.hasYAMLTextAreaDOM) {
          this.codeMirror.setValue(data.data)
        } else {
          this.initCodeMirror(data.data)
        }
        this.isGenerateFailed = false;
      },
      err => {
        this.errorMessage = err.error.message;
        this.isGenerateFailed = true;
      }
    )
  }

  createNewOpenfl() {
    this.isCreatedFailed = false;
    this.isCreatedSubmit = true;

    const directorInfo: DirectorModel = {
      chart_uuid: this.form.controls['chart'].get('chart_uuid')?.value,
      deployment_yaml: this.codeMirror.getTextArea().value,
      description: this.form.controls['info'].get('description')?.value,
      endpoint_uuid: this.form.controls['endpoint'].get('endpoint_uuid')?.value,
      federation_uuid: this.openfl_uuid,
      jupyter_password: this.form.controls['jupyter'].get('password')?.value,
      service_type: +this.form.controls['serviceType'].get('serviceType')?.value,
      registry_config: {
        use_registry: this.useRegistry,
        use_registry_secret: this.useRegistrySecret,
        registry: this.useRegistry ? this.registryConfig : "",
        registry_secret_config: {
          server_url: this.useRegistrySecret ? this.form.controls['registry'].get('server_url')?.value?.trim() : "",
          username: this.useRegistrySecret ? this.form.controls['registry'].get('username')?.value?.trim() : "",
          password: this.useRegistrySecret ? this.form.controls['registry'].get('password')?.value?.trim() : "",
        }
      },
      director_server_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('director_server_cert_info_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('director_server_cert_info_uuid')?.value,
      },
      jupyter_client_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('jupyter_client_cert_info_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('jupyter_client_cert_info_uuid')?.value
      },
      name: this.form.controls['info'].get('name')?.value,
      namespace: this.form.controls['namespace'].get('namespace')?.value
    }
    this.openflService.createDirector(this.openfl_uuid, directorInfo)
      .subscribe(
        data => {
          this.isCreatedFailed = false;
          this.router.navigateByUrl('/federation/openfl/' + this.openfl_uuid)
        },
        err => {
          this.errorMessage = err.error.message;
          this.isCreatedFailed = true;
        }
      );
  }
}
