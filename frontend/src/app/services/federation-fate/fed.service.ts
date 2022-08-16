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

import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class FedService {

  constructor(private http: HttpClient) { }

  getFedList() {
    return this.http.get<any>('/federation');
  }

  getFedDetail(uuid:string) {
    return this.http.get<any>('/federation/fate/'+uuid);
  }

  deleteFed(fed_uuid:string): Observable<any> {
    return this.http.delete('/federation/fate/'+fed_uuid);
  }

  deleteParticipant(fed_uuid:string, type:string, uuid: string, forceRemove:boolean): Observable<any> {
    let params = new HttpParams().set('force', forceRemove);
      return this.http.delete('/federation/fate/'+fed_uuid+'/'+type+'/'+uuid , {params: params});
  }

  createFed(infraInfo:any): Observable<any> {
    return this.http.post('/federation/fate/', infraInfo);
  }

  getFedParticipantList(uuid:string) {
    return this.http.get<any>('/federation/fate/'+uuid+'/participant');
  }

  createExchange(fed_uuid:string, exchangeInfo:any): Observable<any> {
    return this.http.post('/federation/fate/'+ fed_uuid +'/exchange', exchangeInfo);
  }

  getExchangeYaml(chart_uuid:string, namespace:string, name: string, service_type: number, registry: string, use_registry: boolean, use_registry_secret: boolean, enable_psp: boolean) {
    let params = new HttpParams()
    .set('chart_uuid', chart_uuid)
    .set('namespace', namespace)
    .set('name', name)
    .set('service_type', service_type)
    .set('registry', registry)
    .set('use_registry', use_registry)
    .set('use_registry_secret', use_registry_secret)
    .set('enable_psp', enable_psp);
    return this.http.get<any>('/federation/fate/exchange/yaml',{params: params});
  }

  checkPartyID(fed_uuid:string, party_id:number): Observable<any> {
    let params = new HttpParams().set('party_id', party_id);
    return this.http.post('/federation/fate/'+ fed_uuid +'/partyID/check', {},{params: params});
  }

  getClusterYaml(federation_uuid:string, chart_uuid:string, party_id:number, namespace:string, name:string, service_type: number, registry: string, use_registry: boolean, use_registry_secret: boolean, enable_persistence: boolean, storage_class: string, enable_psp: boolean) {
    let params = new HttpParams()
    .set('chart_uuid', chart_uuid)
    .set('federation_uuid', federation_uuid)
    .set('party_id', party_id)
    .set('namespace', namespace)
    .set('name', name)
    .set('service_type', service_type)
    .set('registry', registry)
    .set('use_registry', use_registry)
    .set('use_registry_secret', use_registry_secret)
    .set('enable_persistence', enable_persistence)
    .set('storage_class', storage_class)
    .set('enable_psp', enable_psp);
    return this.http.get<any>('/federation/fate/cluster/yaml',{params: params});
  }

  createCluster(fed_uuid:string, clusterInfo:any): Observable<any> {
    return this.http.post('/federation/fate/'+ fed_uuid +'/cluster', clusterInfo);
  }
  
  getClusterInfo (fed_uuid:string, cluster_uuid:string) {
    return this.http.get<any>(`/federation/fate/${fed_uuid}/cluster/${cluster_uuid}`)
  }

  getExchangeInfo (fed_uuid:string, cluster_uuid:string) {
    return this.http.get<any>(`/federation/fate/${fed_uuid}/exchange/${cluster_uuid}`)
  }

  deleteClusterInfo (fed_uuid:string, cluster_uuid:string) {
    return this.http.delete(`/federation/fate/${fed_uuid}/cluster/${cluster_uuid}`)
  }

  deleteExchangeInfo (fed_uuid:string, cluster_uuid:string) {
    return this.http.delete(`/federation/fate/${fed_uuid}/exchange/${cluster_uuid}`)
  }

  createExternalExchange(fed_uuid:string, externalExchange:any): Observable<any> {
    return this.http.post('/federation/fate/'+ fed_uuid +'/exchange/external', externalExchange);
  }

  createExternalCluster(fed_uuid:string, externalCluster:any): Observable<any> {
    return this.http.post('/federation/fate/'+ fed_uuid +'/cluster/external', externalCluster);
  }
}