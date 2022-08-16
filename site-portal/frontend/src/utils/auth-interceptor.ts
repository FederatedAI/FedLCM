import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpEvent, HttpHandler, HttpRequest, HttpErrorResponse, HttpResponse } from '@angular/common/http';
import { Router } from '@angular/router'
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { MessageService } from '../app/components/message/message.service'
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(public router: Router, private $msg: MessageService) {
  }
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
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
        res => res,
        (err: HttpErrorResponse) => {
          if (err.status === 401 && this.router.url.indexOf('/login') === -1) {
            sessionStorage.removeItem('portal-username')
            sessionStorage.removeItem('sitePortal-redirect')
            const url = this.router.url
            if (newReq.url.indexOf('/user/current') !== -1) {
              this.router.navigateByUrl(`/login`)
            } else {
              sessionStorage.setItem('sitePortal-redirect', url)
              this.router.navigate(['/login/'])
            }
          } else if (err.status === 404) {
            this.$msg.warning('serverMessage.default404')
          }
          return err
        }
      )
    )
  }

}