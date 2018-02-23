import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { HttpService } from '../core/http.service';

declare var _:any;

@Injectable()
export class HMCServerService {

    constructor(private http: HttpService,) {
    }

    jsonParser(key,value) {
        console.log("KEY: "+key+" Value: "+value);
        if ( key == 'Port' ||
             key == 'Freq' ||
             key == 'UpdateScanFreq') {
            return parseInt(value);
        }
        if ( key == 'Active'||
             key == 'ManagedSystemsOnly'||
             key == 'HMCAPIDebug') {
            return ( value === "true" || value === true);
        }
        if ( key == 'ExtraTags' ) {
            return  String(value).split(',');
        }
        return value;
    }

    addHMCServerItem(dev) {
        return this.http.post('/api/cfg/hmcserver',JSON.stringify(dev,this.jsonParser))
        .map( (responseData) => responseData.json());

    }

    editHMCServerItem(dev, id) {
        return this.http.put('/api/cfg/hmcserver/'+id,JSON.stringify(dev,this.jsonParser))
        .map( (responseData) => responseData.json());
    }


    getHMCServerItem(filter_s: string) {
        return this.http.get('/api/cfg/hmcserver')
        .map( (responseData) => {
            return responseData.json();
        })
    }

    getHMCServerItemById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.http.get('/api/cfg/hmcserver/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteHMCServerItem(id : string){
      return this.http.get('/api/cfg/hmcserver/checkondel/'+id)
      .map( (responseData) =>
       responseData.json()
      ).map((deleteobject) => {
          console.log("MAP SERVICE",deleteobject);
          let result : any = {'ID' : id};
          _.forEach(deleteobject,function(value,key){
              result[value.TypeDesc] = [];
          });
          _.forEach(deleteobject,function(value,key){
              result[value.TypeDesc].Description=value.Action;
              result[value.TypeDesc].push(value.ObID);
          });
          return result;
      });
    };

    deleteHMCServerItem(id : string) {
        // return an observable
        return this.http.delete('/api/cfg/hmcserver/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };

    testHMCServer(influxserver,hideAlert?) {
        // return an observable
        return this.http.post('/api/cfg/hmcserver/ping/',JSON.stringify(influxserver,this.jsonParser), null, hideAlert)
        .map((responseData) => responseData.json());
      };
}
