import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TokenType, OpenflType, CreateTokenType, OpenflPsotModel,DirectorModel,ResponseModal } from './openfl-model-type'
@Injectable({
  providedIn: 'root'
})
export class OpenflService {

  constructor(private http: HttpClient) {}

  getOpenflFederationDetail (uuid:string) {
    return this.http.get<OpenflType>(`/federation/openfl/${uuid}`);
  }

  deleteOpenflFederation (fed_uuid:string): Observable<CreateTokenType> {
    return this.http.delete<CreateTokenType>(`/federation/openfl/${fed_uuid}`)
  }

  getTokenList(uuid:string): Observable<TokenType[]> {
    return this.http.get<TokenType[]>(`/federation/openfl/${uuid}/token`);
  }

  createTokenInfo (fed_uuid:string, tokenInfo:any): Observable<CreateTokenType> {
    return this.http.post<CreateTokenType>(`/federation/openfl/${fed_uuid}/token`, tokenInfo);
  }

  deleteTokenInfo (fed_uuid:string, uuid:string) {
    return this.http.delete(`/federation/openfl/${fed_uuid}/token/${uuid}`)
  }

  createOpenflFederation(openflInfo:OpenflPsotModel): Observable<any> {
    return this.http.post('/federation/openfl', openflInfo);
  }

  createDirector (openflId:string,direcortInfo:DirectorModel): Observable<ResponseModal>  {
    return this.http.post<ResponseModal>(`/federation/openfl/${openflId}/director`, direcortInfo)
  }

  deleteDirector (fed_uuid:string, director_uuid:string, forceRemove: boolean): Observable<ResponseModal>  {
    return this.http.delete<ResponseModal>(`/federation/openfl/${fed_uuid}/director/${director_uuid}?force=${forceRemove}`)
  }

  getDirectorYaml(
    federation_uuid:string,
    jupyter_password:string, 
    chart_uuid:string, 
    namespace:string, 
    name: string, 
    service_type: number, 
    registry: string, 
    use_registry: boolean, 
    use_registry_secret: boolean,
    enable_psp: boolean
  ): Observable<ResponseModal> {
    let params = new HttpParams()
    .set('chart_uuid', chart_uuid)
    .set('federation_uuid', federation_uuid)
    .set('namespace', namespace)
    .set('name', name)
    .set('service_type', service_type)
    .set('jupyter_password', jupyter_password)
    .set('registry', registry)
    .set('use_registry', use_registry)
    .set('use_registry_secret', use_registry_secret)
    .set('enable_psp', enable_psp);
    return this.http.get<ResponseModal>('/federation/openfl/director/yaml',{params: params});
  }

  getParticipantInfo (fed_uuid:string): Observable<ResponseModal>  {
    return this.http.get<ResponseModal>(`/federation/openfl/${fed_uuid}/participant`);
  }

  getDirectorInfo (fed_uuid: string, director_uuid: string) : Observable<ResponseModal> {
    return this.http.get<ResponseModal>(`/federation/openfl/${fed_uuid}/director/${director_uuid}`);
  }

  getEnvoyInfo (fed_uuid: string, envoy_uuid: string) : Observable<ResponseModal> {
    return this.http.get<ResponseModal>(`/federation/openfl/${fed_uuid}/envoy/${envoy_uuid}`);
  }

  deleteEnvoy (fed_uuid:string, envoy_uuid:string, forceRemove: boolean): Observable<ResponseModal>  {
    return this.http.delete<ResponseModal>(`/federation/openfl/${fed_uuid}/envoy/${envoy_uuid}?force=${forceRemove}`)
  }
}
