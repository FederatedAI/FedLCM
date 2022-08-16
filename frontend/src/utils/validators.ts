import { Validators, AbstractControl, NG_VALIDATORS, FormGroup } from '@angular/forms';

function myIndexOf(str: string, condition: string): number {
  const newStr = str.toLocaleLowerCase()
  const newCondition = condition.toLocaleLowerCase()
  return newStr.indexOf(newCondition)
}

function maxOrMin(value: any, max: number, min: number, defaultValue = null as any, el: groupModel) {
  let arr = [defaultValue, [EmptyValidator, value]]
  if (el.type[0] === 'notRequired') {
    if (value !== null && value !== 'notRequired') {
      arr = [defaultValue, value]
    } else {
      arr = [defaultValue]
    }
    if (max && min) {
      arr = [defaultValue, [value, Validators.maxLength(max), Validators.minLength(min)]]
    } else if (max) {
      arr = [defaultValue, [value, Validators.maxLength(max)]]
    } else if (min) {
      arr = [defaultValue, [value, Validators.minLength(min)]]
    }
  } else {
    arr = [defaultValue, [EmptyValidator]]
    if (max && min) {
      arr = [defaultValue, [EmptyValidator, value, Validators.maxLength(max), Validators.minLength(min)]]
    } else if (max) {
      arr = [defaultValue, [EmptyValidator, value, Validators.maxLength(max)]]
    } else if (min) {
      arr = [defaultValue, [EmptyValidator, value, Validators.minLength(min)]]
    } else {
      if (value) {
        arr = [defaultValue, [EmptyValidator, value]]
      }
    }
  }
  return arr
}
interface groupModel {
  name: string
  type: string[],
  value?: any
  max?: number
  min?: number
}
export function ValidatorGroup(group: groupModel[]) {
  const newGroup: any = {}
  group.forEach(el => {
    el.type.forEach(key => {
      switch (key) {
        case 'number':
          newGroup[el.name] = maxOrMin(NumberValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'word':
          newGroup[el.name] = maxOrMin(WordValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'noSpace':
          newGroup[el.name] = maxOrMin(noSpaceAlphanumericValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'email':
          newGroup[el.name] = maxOrMin(EmailValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'internet':
          newGroup[el.name] = maxOrMin(InternetValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'ip':
          newGroup[el.name] = maxOrMin(IpValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'zero':
          
          newGroup[el.name] = maxOrMin(zeroToHundred, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'require':
          newGroup[el.name] = maxOrMin('', el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'json':
          newGroup[el.name] = maxOrMin(JsonValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'fqdn':
          newGroup[el.name] = maxOrMin(fqdnValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
        case 'number-list':
          newGroup[el.name] = maxOrMin(NumberListValidator, el.max || 0, el.min || 0, el.value || null, el)
          break;
          case 'notRequired':
          newGroup[el.name] = maxOrMin('notRequired', el.max || 0, el.min || 0, el.value || null, el)
          break;
          default:
          if (el.value) {
            newGroup[el.name] = [el.value]
          } else {
            newGroup[el.name] = [null]
          }
          break;
      }
    })
  })
  
  return newGroup
}
// 
export function EmptyValidator(control: AbstractControl): { [key: string]: any } | null {
  const v = control.value
  if (typeof v === 'string' && !v?.trim()) {
    return { emptyMessage: 'validator.empty' }
  } else if (!v) {
    return { emptyMessage: 'validator.empty' }
  }
  return null
}

// email
export function EmailValidator(control: AbstractControl): { [key: string]: any } | null {
  const v = control.value
  if (myIndexOf(v, '@') === -1) {
    return { message: 'validator.email' }
  }
  return null
}
// numbers
export function NumberValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^[0-9]*$/
  const v = control.value
  if (v === '' || v===null) {
    return null
  }
  if (!reg.test(v)) {
    return { message: 'validator.number' }
  }
  return null
}
// Chinese, English, numbers
export function WordValidator(control: AbstractControl): { [key: string]: any } | null {
  // \u4E00-\u9FA5
  const reg = /^[A-Za-z0-9\s\d\-_/]+$/
  const v = control.value

  if (!reg.test(v?.trim())) {
    return { message: 'validator.word' }
  }
  return null
}
// Chinese, English, numbers
export function noSpaceAlphanumericValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^[\u4E00-\u9FA5A-Za-z0-9\d\-_/]+$/
  const v = control.value

  if (!reg.test(v?.trim())) {
    return { message: 'validator.noSpace' }
  }
  return null
}
// English, numbers
export function EgNuValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^[\u4E00-\u9FA5A-Za-z0-9]+$/
  const v = control.value
  if (!reg.test(v?.trim())) {
    return { message: 'validator.engu' }
  }
  return null
}
// internetUrl
export function InternetValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^(http|https):\/\/([\w]+)\S*/
  const v = control.value
  if (v === '' || v === null) {
    return null
  }
  if (!reg.test(v?.trim())) {
    return { message: 'validator.internet' }
  }
  return null
}

// ip
export function IpValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}/g
  const v = control.value
  if (v === '' || v===null) {
    return null
  }
  if (!reg.test(v?.trim())) {
    return { message: 'validator.ip' }
  }
  return null
}

// 0-100
export function zeroToHundred (control: AbstractControl): { [key: string]: any } | null {
  const v = control.value
  if (v === ''|| v===null) {
    return null
  }
  
  if (+v < 0 || +v >100) {
    return { message: 'validator.zeroToHundred' }
  }
  return null
}


// Json
export function JsonValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^\s*\{\s*[A-Z0-9._]+\s*:\s*[A-Z0-9._]+\s*(,\s*[A-Z0-9._]+\s*:\s*[A-Z0-9._]+\s*)*\}\s*$/i
  const v = control.value
  if (v === '' || v===null || v === '{}') {
    return null
  }
  if (!reg.test(v?.trim())) {
    return { message: 'validator.json' }
  }
  return null
}

// numbers list
export function NumberListValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /^[0-9,]*$/
  const v = control.value
  if (v === '' || v===null) {
    return null
  }
  if (!reg.test(v)) {
    return { message: 'validator.number' }
  }
  return null
}


// FQDN
export function fqdnValidator(control: AbstractControl): { [key: string]: any } | null {
  const reg = /(?=^.{4,253}$)(^((?!-)[a-z0-9-]{0,62}[a-z0-9]\.)+[a-z]{2,63}$)/gm
  const v = control.value
  if (v === '' || v===null) {
    return null
  }
  if (!reg.test(v?.trim())) {
    return { message: 'validator.fqdn' }
  }
  return null
}


//change password matching
export function ConfirmedValidator(controlName: string, matchingControlName: string){

  return (formGroup: FormGroup) => {
      const control = formGroup.controls[controlName];
      const matchingControl = formGroup.controls[matchingControlName];
      if (matchingControl.errors && !matchingControl.errors.confirmedValidator) {
          return;
      }
      if (control.value !== matchingControl.value) {
          matchingControl.setErrors({ confirmedValidator: true });
      } else {
          matchingControl.setErrors(null);
      }

  }

}
