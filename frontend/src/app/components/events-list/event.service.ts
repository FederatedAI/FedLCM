import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';


@Injectable({
  providedIn: 'root'
})
export class EventService {

  constructor(private http: HttpClient) { }
  getEventList(entity_uuid: string) {
    return this.http.get<any>('/event/' + entity_uuid);
  }
}
