interface ConstantModel {
  [key: string]: number
}
interface ResModel {
  name: string
  value: number
}

export const INFRATYPE:ConstantModel = {
  Unknown: 0
}

export const ENDPOINTSTATUS:ConstantModel = {
  Unknown: 0,
  Creating: 1,
  Ready: 2,
  Dismissed: 3,
  Unavailable: 4,
  Deleting: 5
}
export const CHARTTYPE:ConstantModel = {
  Unknown: 0,
  FATEExchange: 1,
  FATECluster: 2,
  OpenFLDirector: 3,
  OpenFLEnvoy: 4
}

export const ParticipantFATEStatus:ConstantModel = {
  Unknown: 0,
  Active: 1,
  Installing: 2,
  Removing: 3,
  Reconfiguring: 4,
  Failed: 5,
  Upgrading: 6
}
export const ParticipantFATEType :ConstantModel = {
  Unknown: 0,
  Exchange: 1,
  Cluster: 2
}

export const CerificateServiceType :ConstantModel = {
  Unknown: 0,
  ATS: 1,
  PulsarServer: 2,
	FMLManagerServer: 3,
	FMLManagerClient: 4,
	SitePortalServer: 5,
	SitePortalClient: 6,
  OpenFLDirector: 101,
	OpenFLJupyter: 102,
	OpenFLEnvoy: 103
}

export const CAType :ConstantModel = {
  Unknown: 0,
  StepCA: 1,
}
export const BindType: ConstantModel= {
  Unknown: 0,
  skip:1,
  existing: 2,
  new: 3
}

export const ServiceType: ConstantModel= {
  Unknown: 0,
  LoadBalancer: 1,
  NodePort: 2
}

export const EventType: ConstantModel= {
  Unknown: 0,
  LogMessage: 1
}

export const CaStatus: ConstantModel= {
  Unknown: 0,
  Unhealthy: 1,
  Healthy: 2
}

export const EnvoyStatus: ConstantModel= {
  Unknown: 0,
  Active: 1,
  Removing: 2,
  Failed: 3,
  InstallingDirector: 4,
  ConfiguringInfra: 5,
  InstallingEndpoint: 6,
  InstallingEnvoy: 7
}

export const Director: ConstantModel= {
  Unknown: 0,
  Active: 1,
  Removing: 2,
  Failed: 3,
  InstallingDirector: 4,
  ConfiguringInfra: 5,
  InstallingEndpoint: 6,
  InstallingEnvoy: 7
}


export function constantGather (inter: string, state: number) {
  let res: ResModel
  switch (inter) {
    case 'infratype':
      res = forObj(INFRATYPE, state)
      break;
    case 'endpointstatus':
      res = forObj(ENDPOINTSTATUS, state)
      break;
    case 'charttype':
      res = forObj(CHARTTYPE, state)
      break;
    case 'participantFATEstatus':
      res = forObj(ParticipantFATEStatus, state)
      break;
    case 'participantFATEtype':
        res = forObj(ParticipantFATEType, state)
      break;
    case 'cerificateType':
        res = forObj(CerificateServiceType, state)
      break;
    case 'caType':
        res = forObj(CAType, state)
      break;
    case 'bindType':
        res = forObj(BindType, state)
      break;
    case 'eventType':
      res = forObj(EventType, state)
    break;
    case 'caStatus':
      res = forObj(CaStatus, state)
    break;
    case 'envoy':
      res = forObj(EnvoyStatus, state)
    break
    case 'director':
      res = forObj(Director, state)
    break
    default:
    res = {
      name: '',
      value: 0
    }
    break;
  }
  return res
}

function forObj (obj: any, state: number): ResModel {
  let res: ResModel = {
    name: '',
    value: 0
  }
  for (const key in obj) {
    if (obj[key] === state) {
      res = {
        name: key,
        value: state
      }
      break
    }
  }
  return res
}