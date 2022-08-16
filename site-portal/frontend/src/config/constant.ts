interface ConstantModel {
  [key: string]: number
}
interface ResModel {
  name: string
  value: number
}
export const JOBTYPE: ConstantModel = {
  Modeling: 1,
  Predict: 2,
  PSI: 3,
  HomoLogisticRegression: 1,
  HomoSecureBoost: 2
}
export const JOBTRAININGTYPE: ConstantModel = {
  HomoLogisticRegression: 1,
  HomoSecureBoost: 2
}
export const JOBSTATUS: ConstantModel = {
  Pending: 1,
  Rejected: 2,
  Running: 3,
  Failed: 4,
  Succeeded: 5,
  Deploying: 6
}
export const PARTYSTATUS: ConstantModel = {
  Unknown: 0,
  Owner: 1,
  Pending: 2,
  Joined: 3,
  Rejected: 4,
  Left: 5,
  Dismissed: 6,
  Revoked: 7
}
export const MODELDEPLOYTYPE: ConstantModel = {
  Unknown: 0,
  KFserving: 1
}
export function constantGather(inter: string, state: number) {
  let res: ResModel
  switch (inter) {
    case 'jobtype':
      res = forObj(JOBTYPE, state)
      break;
    case 'jobstatus':
      res = forObj(JOBSTATUS, state)
      break;
    case 'jobtrainingtype':
      res = forObj(JOBTRAININGTYPE, state)
      break;
    case 'partyStatus':
      res = forObj(PARTYSTATUS, state)
      break;
    default:
      res = {
        name: '',
        value: 0
      }
      break;
    case 'modeldeploytype':
      res = forObj(MODELDEPLOYTYPE, state)
      break;
  }
  return res
}

function forObj(obj: any, state: number): ResModel {
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