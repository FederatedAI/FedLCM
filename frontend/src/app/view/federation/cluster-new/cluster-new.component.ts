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
import { FormGroup, FormBuilder } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { EndpointService } from 'src/app/services/common/endpoint.service';
import { FedService } from 'src/app/services/federation-fate/fed.service';
import { ENDPOINTSTATUS, CHARTTYPE, constantGather } from 'src/utils/constant';
import { ChartService } from 'src/app/services/common/chart.service';
import { EndpointType } from 'src/app/view/endpoint/endpoint-model'
import { ValidatorGroup } from 'src/utils/validators'
import { InfraService } from 'src/app/services/common/infra.service';

@Component({
  selector: 'app-cluster-new',
  templateUrl: './cluster-new.component.html',
  styleUrls: ['./cluster-new.component.scss']
})
export class ClusterNewComponent implements OnInit {
  form: FormGroup;
  constructor(private formBuilder: FormBuilder, private fedservice: FedService, private router: Router, private route: ActivatedRoute, private endpointService: EndpointService, private chartservice: ChartService, private infraservice: InfraService) {
    //form contains configuration to create a new cluster
    this.form = this.formBuilder.group({
      info: this.formBuilder.group({
        name: [''],
        description: [''],
        clusterType: [''],
      }),
      external: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'party_id',
            type: ['number'],
            value: null
          },
          {
            name: 'pulsarHost',
            type: ['ip'],
            value: ""
          },
          {
            name: 'pulsarPort',
            type: ['number'],
            value: 0
          },
          {
            name: 'nginxHost',
            type: ['ip'],
            value: ""
          },
          {
            name: 'nginxPort',
            type: ['number'],
            value: 0
          }
        ])
      ),
      endpoint: this.formBuilder.group({
        endpoint_uuid: ['']
      }),
      chart: this.formBuilder.group({
        chart_uuid: [''],
      }),
      namespace: this.formBuilder.group({
        namespace: [''],
      }),
      party: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'party_id',
            type: ['number'],
            value: null
          }
        ])
      ),
      certificate: this.formBuilder.group({
        cert: ['use'],
        site_portal_client_cert_mode: ['1'],
        site_portal_client_cert_uuid: [''],
        site_portal_client_cert_mode_radio: {value: 'new'},

        site_portal_server_cert_mode: ['1'],
        site_portal_server_cert_uuid: [''],
        site_portal_server_cert_mode_radio: {value: 'new'},
        pulsar_server_cert_info: ['1'],
        pulsar_server_cert_uuid: [''],
        pulsar_server_cert_info_radio: {value: 'new'},
      }),
      serviceType: this.formBuilder.group({
        serviceType: [null],
      }),
      persistence: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'enablePersistence',
            type: ['notRequired'],
            value: false
          },
          {
            name: 'storageClassName',
            type: ['notRequired','noSpace'],
            value: ""
          },
        ])
      ),
      registry: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'useRegistry',
            type: ['notRequired'],
            value: false
          },
          {
            name: 'useRegistrySecret',
            type: ['notRequired'],
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
      externalSpark: this.formBuilder.group(ValidatorGroup([
        {
          name: 'enable_external_spark',
          type: ['notRequired'],
          value: false
        },
        {
          name: 'enable_external_hdfs',
          type: ['notRequired'],
          value: false
        },
        {
          name: 'enable_external_pulsar',
          type: [''],
          value: false
        },
        {
          name: 'external_spark_cores_per_node',
          type: ['notRequired', 'atLeastOne'],
          value: ''
        },
        {
          name: 'external_spark_node',
          type: ['notRequired','atLeastOne'],
          value: ''
        },
        {
          name: 'external_spark_master',
          type: ['notRequired','not-compliant'],
          value: ""
        },
        {
          name: 'external_spark_driverHost',
          type: ['notRequired','ip'],
          value: ""
        },
        {
          name: 'external_spark_driverHostType',
          type: [''],
          value: "NodePort"
        },
        {
          name: 'external_spark_portMaxRetries',
          type: ['notRequired','atLeastOne'],
          value: ''
        },
        {
          name: 'external_spark_driverStartPort',
          type: ['notRequired','numberPort'],
          value: ''
        },
        {
          name: 'external_spark_blockManagerStartPort',
          type: ['notRequired','numberPort'],
          value: ''
        },
        {
          name: 'external_spark_pysparkPython',
          type: [''],
          value: ""
        },
        {
          name: 'external_hdfs_name_node',
          type: ['notRequired','not-compliant'],
          value: ""
        },
        {
          name: 'external_hdfs_path_prefix',
          type: [''],
          value: ""
        },
        {
          name: 'external_pulsar_host',
          type: ['notRequired', 'ip'],
          value: ""
        },
        {
          name: 'external_pulsar_mng_port',
          type: ['notRequired','numberPort'],
          value: ''
        },
        {
          name: 'external_pulsar_port',
          type: ['notRequired','numberPort'],
          value: ''
        },
        {
          name: 'external_pulsar_ssl_port',
          type: ['notRequired','numberPort'],
          value: ''
        }
      ])),

      yaml: this.formBuilder.group({
        yaml: [''],
      })
    });

    this.showEndpointList();
    this.showChartList();
    
  }

  ngOnInit(): void {
  }

  //support to create two type of exchange: create a new one or add an cluster
  get clusterType() {
    return this.form.get('info')?.get('clusterType')?.value
  }
  //isNewCluster returns true when user select to 'create a new one'
  get isNewCluster() {
    return this.form.get('info')?.get('clusterType')?.value === "new"
  }
  //isExternalCluster returns true when user select to 'Add an external one'
  get isExternalCluster() {
    return this.form.get('info')?.get('clusterType')?.value === "external"
  }

  //submitExternalClusterDisable returns true when the input provided for adding an external Cluster is invaid
  get submitExternalClusterDisable() {
    return !this.form.controls['external'].valid || (this.isCreatedExternalSubmit && !this.isCreatedExternalFailed)
  }
  // set Namespace Disabled 
  get setNamespaceDisabled() {
    if (this.selectedEndpoint && this.selectedEndpoint.namespace) {
      this.form.get('namespace')?.get('namespace')?.setValue(this.selectedEndpoint.namespace)
      return true
    } else {
      return false
    }
  }

  // external_spark_disabled returns true when the input provided for adding external spark is invalid
  get external_spark_disabled() {
    const result = []
    result[0] = this.form.get('externalSpark')?.get('enable_external_spark')?.value
    result[1] = this.form.get('externalSpark')?.get('enable_external_hdfs')?.value
    result[2] = this.form.get('externalSpark')?.get('enable_external_pulsar')?.value
    if (result.every(item => item == false)) {
      return false
    } else {
      const arg: string[] = []
      result.forEach((res, index) => {
        if (res) {
          switch (index) {
            case 0:
              arg.push('Spark')
              break;
            case 1:
              arg.push('Hdfs')
              break;
            case 2:
              arg.push('Pulsar')
              break;
                
            default:
              break;
          }
        }
      });      
      return this.externalSparkExtraction(arg)
    }
  }
  // Extraction function to determine whether the Select External Spark next button is disabled
  externalSparkExtraction (arg: string[]) {
    const key1 = arg[0]
    const key2 = arg[1]
    const key3 = arg[2]
    const {sparkValuess, hdfsValues, pulsarValues} = this.extractSparkHandler(false)
    const result: any = {
      resultSpark: false,
      resultHdfs: false,
      resultPulsar: false
    }
    result.resultSpark = sparkValuess.every(item => item.value !== '' && item.value !== null)
    result.resultHdfs = hdfsValues.every(item => item.value !== '' && item.value !== null)
    result.resultPulsar = pulsarValues.every(item => item.value !== '' && item.value !== null)
    
    if (key3) {
      if ((!result['result' + key1]) || (!result['result' + key2]) || (!result['result' + key3])) {
        return true
      }
    } else if (key2) {
      if ((!result['result' + key1]) || (!result['result' + key2])) {
        return true
      }
    } else {
      if (!(result['result' + key1])) {
        return true
      }
    }
    return false
  }

  // Extract spark related word break values.
  extractSparkHandler(bool: boolean) {
    const sparkValuess = []
    sparkValuess[0] = {
      key: 'external_spark_node',
      value: this.form.get('externalSpark')?.get('external_spark_node')?.value
    }
    sparkValuess[1] = {
      key: 'external_spark_master',
      value: this.form.get('externalSpark')?.get('external_spark_master')?.value
    }
    sparkValuess[2] = {
      key: 'external_spark_cores_per_node',
      value: this.form.get('externalSpark')?.get('external_spark_cores_per_node')?.value
    }
    sparkValuess[3] = {
      key: 'external_spark_driverHostType',
      value: this.form.get('externalSpark')?.get('external_spark_driverHostType')?.value
    }
    sparkValuess[4] = {
      key: 'external_spark_portMaxRetries',
      value: this.form.get('externalSpark')?.get('external_spark_portMaxRetries')?.value
    }
    sparkValuess[5] = {
      key: 'external_spark_driverStartPort',
      value: this.form.get('externalSpark')?.get('external_spark_driverStartPort')?.value
    }
    sparkValuess[6] = {
      key: 'external_spark_blockManagerStartPort',
      value: this.form.get('externalSpark')?.get('external_spark_blockManagerStartPort')?.value
    }
    sparkValuess[7] = {
      key: 'external_spark_driverHost',
      value: this.form.get('externalSpark')?.get('external_spark_driverHost')?.value
    }
    const hdfsValues = []
    const pulsarValues = []

    pulsarValues[0] = {
      key: 'external_pulsar_mng_port',
      value: this.form.get('externalSpark')?.get('external_pulsar_mng_port')?.value
    }
    pulsarValues[1] = {
      key: 'external_pulsar_port',
      value: this.form.get('externalSpark')?.get('external_pulsar_port')?.value
    }
    pulsarValues[2] = {
      key: 'external_pulsar_ssl_port',
      value: this.form.get('externalSpark')?.get('external_pulsar_ssl_port')?.value
    }
    pulsarValues[3] = {
      key: 'external_pulsar_host',
      value: this.form.get('externalSpark')?.get('external_pulsar_host')?.value
    }

    if (bool) {
      sparkValuess[8] = {
        key: 'external_spark_pysparkPython',
        value: this.form.get('externalSpark')?.get('external_spark_pysparkPython')?.value
      }
      hdfsValues[0] = {
        key: 'external_hdfs_name_node',
        value: this.form.get('externalSpark')?.get('external_hdfs_name_node')?.value
      }
      hdfsValues[1] = {
        key: 'external_hdfs_path_prefix',
        value: this.form.get('externalSpark')?.get('external_hdfs_path_prefix')?.value
      }

    }
    return {sparkValuess, hdfsValues, pulsarValues}
  }

  isCreatedExternalSubmit = false;
  isCreatedExternalFailed = false;
  //createExternalCluster is to submit the request to 'create an external cluster'
  createExternalCluster() {
    this.isCreatedExternalSubmit = true;
    this.isCreatedExternalFailed = false;
    var externalCluster = {
      name: this.form.controls['info'].get('name')?.value,
      description: this.form.controls['info'].get('description')?.value,
      federation_uuid: this.fed_uuid,
      party_id: Number(this.form.get('external')?.get('party_id')?.value?.trim()),
      pulsar_access_info: {
        host: this.form.get('external')?.get('pulsarHost')?.value?.trim(),
        port: Number(this.form.get('external')?.get('pulsarPort')?.value?.trim()),
        fqdn: this.form.get('external')?.get('pulsarFQDN')?.value?.trim()
      },
      nginx_access_info: {
        host: this.form.get('external')?.get('nginxHost')?.value?.trim(),
        port: Number(this.form.get('external')?.get('nginxPort')?.value?.trim()),
      }
    }
    if (this.form.controls['external'].valid && this.form.controls['info'].valid) {
      this.fedservice.createExternalCluster(this.fed_uuid, externalCluster)
        .subscribe(
          data => {
            this.isCreatedExternalFailed = false;
            this.router.navigateByUrl('/federation/fate/' + this.fed_uuid)
          },
          err => {
            this.errorMessage = err.error.message;
            this.isCreatedExternalFailed = true;
          }
        );
    } else {
      this.errorMessage = "invalid input";
      this.isCreatedExternalFailed = true;
    }
    return
  }

  endpointStatus = ENDPOINTSTATUS;
  constantGather = constantGather
  endpointlist: any = [];
  errorMessage: any;
  isShowEndpointFailed: boolean = false;
  isPageLoading: boolean = true;
  noEndpoint = false;

  //showEndpointList is to get the ready endpoint list
  showEndpointList() {
    this.isShowEndpointFailed = false;
    this.isPageLoading = true;
    this.endpointlist = []
    this.endpointService.getEndpointList()
      .subscribe((data: any) => {
        if (data.data) {
          for (var ep of data.data) {
            if (ep?.status === 2) this.endpointlist?.push(ep)
          }
          if (this.endpointlist.length === 0) {
            this.noEndpoint = true;
            this.isPageLoading = false;
            return
          }
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowEndpointFailed = true;
          this.isPageLoading = false;
        });
  }

  fed_uuid = String(this.route.snapshot.paramMap.get('id'));
  openInfraModal: boolean = false;

  endpoint!: EndpointType | null
  //selectedEndpoint is the selected endpoint in the 'Step 2'
  get selectedEndpoint() {
    return this.endpoint
  }
  set selectedEndpoint(value) {
    this.endpoint = value
    if (value) {
      this.onSelectEndpoint()
    }
  }

  use_cert: boolean = true;
  skip_cert: boolean = false;
  //setRadioDisplay is trigger when the selection of certificate's mode is changed
  setRadioDisplay(val: any) {
    if (val == "use") {
      this.use_cert = true;
      this.skip_cert = false;
    }
    if (val == "skip") {
      this.use_cert = false;
      this.skip_cert = true;
      this.form.get('certificate')?.get('site_portal_client_cert_mode')?.setValue(1);
      this.form.get('certificate')?.get('site_portal_server_cert_mode')?.setValue(1);
      this.form.get('certificate')?.get('pulsar_server_cert_info')?.setValue(1);
    }
  }

  //cert_disabled is to validate the configuration of certificate section and disabled the 'Next' button if needed
  get cert_disabled() {
    // cert is skip
    if (this.form.get('certificate')?.get('cert')?.value === 'skip') {
      return false
    } else {
      // isChartContainsPortalservices value is true
      if (this.isChartContainsPortalservices) {
        const pulsar_server_cert_info_radio = this.form.get('certificate')?.get('pulsar_server_cert_info_radio')?.value
        const site_portal_server_cert_mode_radio = this.form.get('certificate')?.get('site_portal_server_cert_mode_radio')?.value
        const site_portal_client_cert_mode_radio = this.form.get('certificate')?.get('site_portal_client_cert_mode_radio')?.value
        if (pulsar_server_cert_info_radio && site_portal_server_cert_mode_radio && site_portal_client_cert_mode_radio) {
          return false
        } else {
          return true
        }
      } else {
        const pulsar_server_cert_info_radio = this.form.get('certificate')?.get('pulsar_server_cert_info_radio')?.value
        if (pulsar_server_cert_info_radio) {
          return false
        } else {
          return true
        }
      }
    }

  }

  //service_type_disabled is to validate the configuration of Service type section and disabled the 'Next' button if needed
  get service_type_disabled() {
    return !this.form.controls['serviceType'].get('serviceType')?.valid
  }

  //reset form when selection change
  endpointSelectOk: boolean = false;
  infraUUID: any = ""
  onSelectEndpoint() {
    if (this.selectedEndpoint) {
      this.endpointSelectOk = true;
      this.form.get('endpoint')?.get('endpoint_uuid')?.setValue(this.selectedEndpoint.uuid);
      this.infraUUID = this.endpoint?.infra_provider_uuid
      this.showInfraDetail(this.infraUUID)
    }
  }

  infraConfigDetail: any;
  isShowInfraDetailFailed: boolean = false;
  hasRegistry = false;
  hasRegistrySecret = false;
  registrySecretConfig: any;
  //showInfraDetail is to get the registry/registry secret information that saved on the infra of the selected endpoint
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

  get useRegistrySecret() {
    return this.form.controls['registry'].get('useRegistrySecret')?.value;
  }
  get useRegistry() {
    return this.form.controls['registry'].get('useRegistry')?.value;
  }
  get registryConfig() {
    return this.form.controls['registry'].get('registryConfig')?.value;
  }

  //onChange_use_registry is triggered when the value of 'useRegistry' is changed
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

  //onChange_use_registry_secre is triggered when the value of 'useRegistrySecret' is changed
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

  //registry_disabled is to valid if the registry/registry secret configuration are invalid
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

  //persistence_disabled is to valid if the persistence configuration if invalid
  get persistence_disabled() {
    return !((this.form.controls['persistence'].get('enablePersistence')?.value && this.form.get('persistence.storageClassName')?.valid) || !this.form.controls['persistence'].get('enablePersistence')?.value)
  }

  //enablePersistence returns if enable the Persistent volume
  get enablePersistence() {
    if (this.form.controls['persistence'].get('enablePersistence')?.value === null) {
      this.form.get('persistence')?.get('storageClassName')?.setValue("-")
      return false
    }
    return this.form.controls['persistence'].get('enablePersistence')?.value
  }

  //onChange_enable_persistence is triggered when the value of 'enable_persistence' is changed
  onChange_enable_persistence(val: any) {
    this.selectChange(val)
    this.change_enable_persistence()
  }
  change_enable_persistence() {
    if (this.enablePersistence) {
      this.form.get('persistence')?.get('storageClassName')?.setValue("");
    } else {
      this.form.get('persistence')?.get('storageClassName')?.setValue("-");
    }
  }

  //formResetChild is to reset group form, child is the group name in the form
  formResetChild(child: string) {
    //need to reset twice due to the issues of Clarity Steppers
    this.form.controls[child].reset();
    this.form.controls[child].reset();
    if (child === 'yaml') {
      this.codeMirror?.setValue('')
    }
  }

  partyId_valid = false;
  isCheckSuccess = false;
  isCheckFailed = false;
  isLoading = false;
  //checkPartyID is to check the confict of Party ID
  checkPartyID() {
    this.isCheckSuccess = false;
    this.isCheckFailed = false;
    this.isLoading = true;
    if (this.form.controls['party'].get('party_id')?.value !== null) {
      const party_id = Number(this.form.controls['party'].get('party_id')?.value);
      this.fedservice.checkPartyID(this.fed_uuid, party_id).subscribe(
        data => {
          this.partyId_valid = true;
          this.form.get('namespace')?.get('namespace')?.setValue('fate-' + party_id);
          this.isLoading = false;
          this.isCheckSuccess = true;
        },
        err => {
          this.errorMessage = err.error.message;
          this.partyId_valid = false;
          this.isLoading = false;
          this.isCheckFailed = true;
        }
      )
    } else {
      this.partyId_valid = false;
      this.isLoading = false;
      this.isCheckFailed = true;
      this.errorMessage = "invalid party ID";
    }
  }

  //editPartyID is to enable the form to edit and reset the value of yaml
  editPartyID() {
    this.partyId_valid = false;
    this.isCheckSuccess = false;
    this.isCheckFailed = false;
    this.isLoading = false;
    this.formResetChild('yaml');
    this.formResetChild('party');
    this.isGenerateSubmit = false;
    this.isGenerateFailed = false;
  }

  //selectChange is to reset YAML value when the releted confiuration is changed
  selectChange(val: any) {
    this.formResetChild('yaml');
    this.isGenerateSubmit = false;
    this.isGenerateFailed = false;
  }

  //onChartChange is triggered when the selection of chart is changed
  isChartContainsPortalservices = false;
  onChartChange(val: any) {
    var chart_uuid = this.form.controls['chart'].get('chart_uuid')?.value;
    this.chartservice.getChartDetail(chart_uuid)
      .subscribe((data: any) => {
        if (data.data) {
          //isChartContainsPortalservices to decide the certificate we need when creating the exchange
          this.isChartContainsPortalservices = data.data.contain_portal_services;
        }
        this.isPageLoading = false;
        this.isShowChartFailed = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
    this.selectChange(val);
    this.resetCert();
  }

  chartType = CHARTTYPE;
  chartlist: any = [];
  isShowChartFailed: boolean = false;
  //showChartList is to get the chart list of deloying exchange
  showChartList() {
    this.chartlist = [];
    this.isPageLoading = true;
    this.isShowChartFailed = false;
    this.chartservice.getChartList()
      .subscribe((data: any) => {
        if (data.data) {
          for (let chart of data.data) {
            if (chart.type === 2) {
              this.chartlist.push(chart);
            }
          }
        }
        this.isPageLoading = false;
        this.isShowChartFailed = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
  }

  //resetCert is triggered when the selection of chart is changed
  resetCert() {
    this.formResetChild('certificate');
    this.form.get('certificate')?.get('cert')?.setValue("use");
    this.setRadioDisplay("use");
  }

  //enablePSP returns if enable the pod secrurity policy
  get enablePSP() {
    return this.form.controls['psp'].get('enablePSP')?.value;
  }

  isGenerateSubmit = false;
  isGenerateFailed = false;
  codeMirror: any
  //generateYaml is to get the exchange initial yaml based on the configuration provided by user
  generateClusterYaml() {
    this.isGenerateSubmit = true;
    const spark = this.form.get('externalSpark')?.get('enable_external_spark')?.value
    const hdfs = this.form.get('externalSpark')?.get('enable_external_hdfs')?.value
    const pulsar = this.form.get('externalSpark')?.get('enable_external_pulsar')?.value
    const {sparkValuess, hdfsValues, pulsarValues} = this.extractSparkHandler(true)
    // Build the passed query parameter list
    const queryList: any = [
      {
        key: 'federation_uuid',
        value: this.fed_uuid
      },
      {
        key: 'chart_uuid',
        value: this.form.controls['chart'].get('chart_uuid')?.value
      },
      {
        key: 'party_id',
        value: this.form.controls['party'].get('party_id')?.value
      },
      {
        key: 'namespace',
        value: this.form.controls['namespace'].get('namespace')?.value?.trim()
      },
      {
        key: 'name',
        value: this.form.controls['info'].get('name')?.value?.trim()
      },
      {
        key: 'service_type',
        value: Number(this.form.controls['serviceType'].get('serviceType')?.value)
      },
      {
        key: 'registry',
        value: this.registryConfig
      },
      {
        key: 'use_registry',
        value: this.useRegistry
      },
      {
        key: 'use_registry_secret',
        value: this.useRegistrySecret
      },
      {
        key: 'enable_persistence',
        value: this.form.controls['persistence'].get('enablePersistence')?.value
      },
      {
        key: 'storage_class',
        value: this.enablePersistence ? this.form.controls['persistence'].get('storageClassName')?.value?.trim() : ""
      },
      {
        key: 'enable_psp',
        value: this.enablePSP
      },
      {
        key: 'enable_external_spark',
        value: spark
      },
      {
        key: 'enable_external_hdfs',
        value: hdfs
      },
      {
        key: 'enable_external_pulsar',
        value: pulsar
      }
    ]
    if (spark) {
      sparkValuess.forEach(item => {
        queryList.push(item)
      })
    }
    if (hdfs) {
      hdfsValues.forEach(item => {
        queryList.push(item)
      })
    }
    if (pulsar) {
      pulsarValues.forEach(item => {
        queryList.push(item)
      })
    }

    this.fedservice.getClusterYaml(queryList).
    subscribe(

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

  //submitDisable returns if disabled the submit button of create a new cluster
  get submitDisable() {
    if (!this.form.controls['info'].valid || this.cert_disabled || this.service_type_disabled || !this.isGenerateSubmit || (this.isCreatedSubmit && !this.isCreatedFailed)) {
      return true
    } else {
      return false
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

  isCreatedSubmit = false;
  isCreatedFailed = false;
  //createNewCluster is to submmit the request of 'create a new cluster'
  createNewCluster() {
    this.isCreatedFailed = false;
    this.isCreatedSubmit = true;
    if (this.isChartContainsPortalservices) {
      if(this.use_cert) {
        this.form.controls['certificate'].get('pulsar_server_cert_info')?.setValue('3')
        this.form.controls['certificate'].get('site_portal_client_cert_mode')?.setValue('3')
        this.form.controls['certificate'].get('site_portal_server_cert_mode')?.setValue('3')
      } else {
        this.form.controls['certificate'].get('pulsar_server_cert_info')?.setValue('1')
        this.form.controls['certificate'].get('site_portal_client_cert_mode')?.setValue('1')
        this.form.controls['certificate'].get('site_portal_server_cert_mode')?.setValue('1')
      }
    } else {
      if(this.use_cert) {
        this.form.controls['certificate'].get('pulsar_server_cert_info')?.setValue('3')
        this.form.controls['certificate'].get('site_portal_client_cert_mode')?.setValue('1')
        this.form.controls['certificate'].get('site_portal_server_cert_mode')?.setValue('1')
      } else {
        this.form.controls['certificate'].get('pulsar_server_cert_info')?.setValue('1')
        this.form.controls['certificate'].get('site_portal_client_cert_mode')?.setValue('1')
        this.form.controls['certificate'].get('site_portal_server_cert_mode')?.setValue('1')
      }
    }

    const clusterInfo = {
      chart_uuid: this.form.controls['chart'].get('chart_uuid')?.value,
      deployment_yaml: this.codeMirror.getTextArea().value,
      description: this.form.controls['info'].get('description')?.value,
      endpoint_uuid: this.form.controls['endpoint'].get('endpoint_uuid')?.value,
      federation_uuid: this.fed_uuid,
      name: this.form.controls['info'].get('name')?.value,
      namespace: this.form.controls['namespace'].get('namespace')?.value ? this.form.controls['namespace'].get('namespace')?.value : 'fate-' + this.form.controls['party'].get('party_id')?.value,
      party_id: Number(this.form.controls['party'].get('party_id')?.value),
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
      pulsar_server_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('pulsar_server_cert_info')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('pulsar_server_cert_uuid')?.value
      },
      site_portal_client_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('site_portal_client_cert_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('site_portal_client_cert_uuid')?.value
      },
      site_portal_server_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('site_portal_server_cert_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('site_portal_server_cert_uuid')?.value
      }
    }


    this.fedservice.createCluster(this.fed_uuid, clusterInfo)
      .subscribe(
        data => {
          this.isCreatedFailed = false;
          this.router.navigateByUrl('/federation/fate/' + this.fed_uuid)
        },
        err => {
          this.errorMessage = err.error.message;
          this.isCreatedFailed = true;
        }
      );
  }

  //validURL is to validate URL is valid
  validURL(str: string) {
    var pattern = new RegExp(
      '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|' + // domain name
      '((\\d{1,3}\\.){3}\\d{1,3}))' + // OR ip (v4) address
      '(\\?[;&a-z\\d%_.~+=-]*)?' + // query string
      '(\\#[-a-z\\d_]*)?$', 'i'); // fragment locator
    return !!pattern.test(str);
  }

  //server_url_suggestion to return the suggested server url based on the provided registry
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

  //valid_server_url return if the the server url is valid or not
  get valid_server_url() {
    return this.form.get('registry.server_url')?.valid && this.form.controls['registry'].get('server_url')?.value?.trim() != ''
  }

  // onChange_use_registry fires when the value of the incoming 'type' changes
  onChange_external_spark(data: Boolean, type: 'spark' | 'hdfs' | 'pulsar') {

  }

}

