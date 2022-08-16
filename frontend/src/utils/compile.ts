export function compile (code: string) {
  let newStr = String.fromCharCode(code.charCodeAt(0) + code.length)
  for (let i = 1; i < code.length; i++){
    newStr += String.fromCharCode(code.charCodeAt(i) + code.length)
  }
  return escape(newStr)
}

export function uncompile (code: string) {
  code = unescape(code)
  
  let newStr = String.fromCharCode(code.charCodeAt(0) - code.length)
  for (let i = 1; i < code.length; i++) {
    newStr += String.fromCharCode(code.charCodeAt(i) - code.length)
  }
  return newStr
}