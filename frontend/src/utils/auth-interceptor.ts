import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpEvent, HttpHandler, HttpRequest, HttpErrorResponse, HttpResponse } from '@angular/common/http';
import { Router } from '@angular/router'
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { MessageService } from '../app/components/message/message.service'
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor (public router: Router,private $msg: MessageService) {
  }
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    let operationSet = new Set<string>(['connect', 'scan', 'check', 'login', 'logout', 'upload']);
    let newReq: HttpRequest<any>
    // get language package
    if (req.url.indexOf('../assets') === -1) {
      newReq = req.clone({
        url: '/api/v1' + req.url
      })
    } else {// other request
      newReq = req.clone()
    }
    return next.handle(newReq).pipe(
      tap(
        (res:any) => {          
          if (res.status === 200 && req.method!== 'GET') {
            const str = req.url.split('/')
            this.$msg.close()
            if (req.method === 'POST') {
              var lastWord = str[str.length-1]
              if (operationSet.has(lastWord)) {   
                if (lastWord !== 'login') {
                  this.$msg.success('serverMessage.'+ str[str.length-1], 2000)
                }             
              } else {
                this.$msg.success('serverMessage.create200',2000)
              }
            } else if (req.method === 'PUT') {
              this.$msg.success('serverMessage.update200',2000)
            } else if (req.method === 'DELETE') {
              this.$msg.success('serverMessage.delete200',2000)
            }   
          }
          return res
        },
        (err: HttpErrorResponse) => {
          if (err.status === 401 && this.router.url.indexOf('/login') === -1) {
            sessionStorage.removeItem('lifecycleManager-redirect')
            const url = this.router.url
            if (newReq.url.indexOf('/user/current')!== -1) {
              this.router.navigateByUrl(`/login`)
            } else {
              if (url !== '/federation' && url !== '/timeout') {
                sessionStorage.setItem('lifecycleManager-redirect', url)
              }
              this.router.navigate(['/login/'])
            }
          } else if (err.status === 404) {
            this.$msg.error('serverMessage.default404')
          } 
          return err
        }
      )
    )
  }
  
}