export interface TokenType {
  "description": string,
  "expired_at": string,
  "labels": {[key:string]:string},
  "limit": number,
  "name": string,
  "token_str": string,
  "uuid": string
}

export interface OpenflType {
  "created_at": string,
  "description": string,
  "domain": string,
  "name": string,
  "shard_descriptor_config": {
    "envoy_config_yaml": string,
    "python_files": {[key:string]:string},
    "sample_shape": string[],
    "target_shape": string[]
  },
  "type": string,
  "use_customized_shard_descriptor": boolean,
  "uuid": string
}

export interface CreateTokenType {
  "description": string,
  "expired_at": string | Date,
  "labels": {[key:string]:string},
  "limit": number,
  "name": string
}

export interface OpenflPsotModel {
  "description": string,
  "domain": string,
  "name": string,
  "shard_descriptor_config": {
    "envoy_config_yaml": string,
    "python_files": {[key:string]: string}
    "sample_shape": string[],
    "target_shape": string[]
  },
  "use_customized_shard_descriptor": boolean
}

export interface DirectorModel {
  "chart_uuid": string,
  "deployment_yaml": string,
  "description": string,
  "director_server_cert_info": {
    "binding_mode": number
    "common_name": string,
    "uuid": string
  },
  "endpoint_uuid": string,
  "federation_uuid": string,
  "jupyter_client_cert_info": {
    "binding_mode": number,
    "common_name": string,
    "uuid": string
  },
  "jupyter_password": string,
  "name": string,
  "namespace": string,
  "registry_config": {
    "registry": string,
    "registry_secret_config": {
      "password": string,
      "server_url": string,
      "username": string
    },
    "use_registry": boolean,
    "use_registry_secret": boolean
  },
  "service_type": number
}
export interface ResponseModal {
    "code": number
    "data": any
    "message": string
}
export interface LabelModel {
  key:string
  value:string
  no?: number
}
export interface EnvoyModel {
  "uuid": string
"name": string
"description": string
"created_at": string
"type": number
"endpoint_name": string
"endpoint_uuid": string
"infra_provider_name": string
"infra_provider_uuid": string
"namespace": string
"cluster_uuid": string
"status": number,
"access_info": any
"token_str": string
"token_name": string
"labels": {[key:string]:string}[],
"selected":boolean,
"deleteFailed":boolean,
"deleteSuccess":boolean,
"deleteSubmit":boolean,
"errorMessage":boolean,
showLabelListFlag?:boolean
}

export interface DirectorInfoModel {
  "access_info": {[key:string]:any}
  "cluster_uuid": string,
  "created_at": string,
  "description": string,
  "endpoint_name": string,
  "endpoint_uuid": string,
  "infra_provider_name": string,
  "infra_provider_uuid": string,
  "name": string,
  "namespace": string,
  "status": number,
  "type": number,
  "uuid": string
}

export interface EnvoyInfoModel {
  "access_info": {
    "additionalProp1": {
      "fqdn": string,
      "host": string,
      "port": number,
      "service_type": string,
      "tls": boolean
    },
    "additionalProp2": {
      "fqdn": string,
      "host": string,
      "port": number,
      "service_type": string,
      "tls": boolean
    },
    "additionalProp3": {
      "fqdn": string,
      "host": string,
      "port": number,
      "service_type": string,
      "tls": boolean
    }
  },
  "cluster_uuid": string,
  "created_at": string,
  "description": string,
  "endpoint_name": string,
  "endpoint_uuid": string,
  "infra_provider_name": string,
  "infra_provider_uuid": string,
  "name": string,
  "namespace": string,
  "status": number,
  "type": number,
  "uuid": string
}