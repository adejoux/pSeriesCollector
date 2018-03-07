import { Injectable } from '@angular/core';
import { HttpService } from '../core/http.service';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class DeviceCfgService {

    constructor(public httpAPI: HttpService) {
    }

    parseJSON(key,value) {
        if ( key == 'NmonFreq'  ||
        key == 'Timeout' ) {
          return parseInt(value);
        }
        if ( key == 'EnableHMCStats' ||
        key == 'EnableNmonStats' ||
        key == 'NmonProtDebug') return ( value === "true" || value === true);
        if ( key == 'ExtraTags' ) {
            return  String(value).split(',');
        }
        return value;
    }

    addDeviceCfg(dev) {
        return this.httpAPI.post('/api/cfg/devices',JSON.stringify(dev,this.parseJSON))
        .map( (responseData) => responseData.json());

    }

    editDeviceCfg(dev, id, hideAlert?) {
        return this.httpAPI.put('/api/cfg/devices/'+id,JSON.stringify(dev,this.parseJSON),null,hideAlert)
        .map( (responseData) => responseData.json());
    }

    getDeviceCfg(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/devices')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((devices) => {
            console.log("MAP SERVICE",devices);
            let result = [];
            if (devices) {
                _.forEach(devices,function(value,key){
                    console.log("FOREACH LOOP",value,value.ID);
                    if(filter_s && filter_s.length > 0 ) {
                        console.log("maching: "+value.ID+ "filter: "+filter_s);
                        var re = new RegExp(filter_s, 'gi');
                        if (value.ID.match(re)){
                            result.push(value);
                        }
                        console.log(value.ID.match(re));
                    } else {
                        result.push(value);
                    }
                });
            }
            return result;
        });
    }
    getDeviceCfgById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/devices/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteDeviceCfg(id : string){
      return this.httpAPI.get('/api/cfg/devices/checkondel/'+id)
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

    testDeviceCfg(influxserver,hideAlert?) {
      // return an observable
      return this.httpAPI.post('/api/cfg/devices/ping/',JSON.stringify(influxserver,this.parseJSON), null, hideAlert)
      .map((responseData) => responseData.json());
    };

    deleteDeviceCfg(id : string, hideAlert?) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/devices/'+id, null, hideAlert)
        .map( (responseData) =>
         responseData.json()
        );
    };
}
