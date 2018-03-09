import { Component, ChangeDetectionStrategy, ViewChild } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';

import { DeviceCfgService } from './devicecfg.service';
import { ValidationService } from '../common/custom-validation/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/custom-modal/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { ItemsPerPageOptions } from '../common/global-constants';
import { TableActions } from '../common/table-actions';
import { AvailableTableActions } from '../common/table-available-actions';

import { TableListComponent } from '../common/table-list.component';
import { DeviceCfgComponentConfig, TableRole, OverrideRoleActions } from './devicecfg.data';
import { InfluxServerService } from '../influxserver/influxservercfg.service';

declare var _:any;

@Component({
  selector: 'devicecfg',
  providers: [DeviceCfgService, ValidationService,InfluxServerService],
  templateUrl: './deviceeditor.html',
  styleUrls: ['../../css/component-styles.css']
})

export class DeviceCfgComponent {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  itemsPerPageOptions : any = ItemsPerPageOptions;
  editmode: string; //list , create, modify
  influxservers: Array<any>;
  filter: string;
  deviceForm: any;
  myFilterValue: any;
  alertHandler : any = null;

  public selectinfluxservers: IMultiSelectOption[] = [];
  private mySettingsInflux: IMultiSelectSettings = {
    singleSelect: true,
  };



  //Initialization data, rows, colunms for Table
  private data: Array<any> = [];
  public rows: Array<any> = [];
  public tableAvailableActions : any;

  selectedArray : any = [];
  public defaultConfig : any = DeviceCfgComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;
  public isRequesting : boolean;
  public counterItems : number = null;
  public counterErrors: any = [];

  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number = 5;
  public numPages: number = 1;
  public length: number = 0;
  private builder;
  private oldID : string;
  
  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.defaultConfig['table-columns'] },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public influxServerService: InfluxServerService,public deviceCfgService: DeviceCfgService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.editmode = 'list';
    this.reloadData();
    this.builder = builder;
  }

  createStaticForm() {
    this.deviceForm = this.builder.group({
      ID: [this.deviceForm ? this.deviceForm.value.ID : '', Validators.required],
      Name: [this.deviceForm ? this.deviceForm.value.Name : '', Validators.required],
      SerialNumber: [this.deviceForm ? this.deviceForm.value.SerialNumber : '', Validators.required],
      OSVersion: [this.deviceForm ? this.deviceForm.value.OSVersion : ''],
      Type: [this.deviceForm ? this.deviceForm.value.Type : ''],
      Location: [this.deviceForm ? this.deviceForm.value.Location : ''],
      EnableHMCStats: [this.deviceForm ? this.deviceForm.value.EnableHMCStats : 'true', Validators.required],
      EnableNmonStats: [this.deviceForm ? this.deviceForm.value.EnableNmonStats : 'true', Validators.required],
      NmonFreq: [this.deviceForm ? this.deviceForm.value.NmonFreq : 60,  ValidationService.uintegerNotZeroValidator],
      NmonOutDB: [this.deviceForm ? this.deviceForm.value.NmonOutDB : '', Validators.required],
      NmonIP: [this.deviceForm ? this.deviceForm.value.NmonIP : ''],
      NmonSSHUser: [this.deviceForm ? this.deviceForm.value.NmonSSHUser : ''],
      NmonSSHKey: [this.deviceForm ? this.deviceForm.value.NmonSSHKey : ''],
      NmonLogLevel: [this.deviceForm ? this.deviceForm.value.NmonLogLevel : 'info'],
      NmonFilePath: [this.deviceForm ? this.deviceForm.value.NmonFilePath : '/var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon'],
      NmonProtDebug: [this.deviceForm ? this.deviceForm.value.NmonProtDebug : 'false'],
      ExtraTags: [this.deviceForm ? (this.deviceForm.value.ExtraTags ? this.deviceForm.value.ExtraTags : "" ) : "" , Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags])],
      Description: [this.deviceForm ? this.deviceForm.value.Description : '']
    });
  }

  reloadData() {
    // now it's a simple subscription to the observable
    this.alertHandler = null;
    this.deviceCfgService.getDeviceCfg(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.influxservers = data
        this.data = data;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  applyAction(test : any, data? : Array<any>) : void {
    this.selectedArray = data || [];
    switch(test.action) {
       case "RemoveAllSelected": {
          this.removeAllSelectedItems(this.selectedArray);
          break;
       }
       case "ChangeProperty": {
          this.updateAllSelectedItems(this.selectedArray,test.field,test.value)
          break;
       }
       case "AppendProperty": {
         this.updateAllSelectedItems(this.selectedArray,test.field,test.value,true);
       }
       default: {
          break;
       }
    }
  }

  customActions(action : any) {
    switch (action.option) {
      case 'reload' :
        this.reloadData();
      break;
      case 'export' : 
        this.exportItem(action.event);
      break;
      case 'new' :
        this.newItem()
      break;
      case 'view':
        this.viewItem(action.event);
      break;
      case 'edit':
        this.editItem(action.event);
      break;
      case 'remove':
        this.removeItem(action.event);
      break;
      case 'tableaction':
        this.applyAction(action.event, action.data);
      break;
      case 'editenabled':
        this.enableEdit();
      break;
    }
  }

  enableEdit() {
    this.influxServerService.getInfluxServer(null)
      .subscribe(
        data => {
          this.tableAvailableActions = new AvailableTableActions(this.defaultConfig['slug'],this.createMultiselectArray(data,'ID','ID','Description')).availableOptions
        },
        err => console.log(err),
        () => console.log()
      );
    }

  viewItem(id) {
    console.log('view', id);
    this.viewModal.parseObject(id);
  }

  exportItem(item : any) : void {
    this.exportFileModal.initExportModal(item);
  }

  removeAllSelectedItems(myArray) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    for (let i in myArray) {
      console.log("Removing ",myArray[i].ID)
      this.deleteDeviceCfg(myArray[i].ID,true);
      obsArray.push(this.deleteDeviceCfg(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.deviceCfgService.checkOnDeleteDeviceCfg(id)
      .subscribe(
      data => {
        console.log(data);
        let temp = data;
        this.viewModalDelete.parseObject(temp)
      },
      err => console.error(err),
      () => { }
      );
  }
  newItem() {
    //No hidden fields, so create fixed Form
    this.createStaticForm();
    this.editmode = "create";
    this.getInfluxServersforDevices()
  }

  editItem(row) {
    let id = row.ID;
    this.getInfluxServersforDevices()
    this.deviceCfgService.getDeviceCfgById(id)
      .subscribe(data => {
        this.deviceForm = {};
        this.deviceForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteDeviceCfg(id, recursive?) {
    if (!recursive) {
    this.deviceCfgService.deleteDeviceCfg(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.editmode = "list"; this.reloadData() }
      );
    } else {
      return this.deviceCfgService.deleteDeviceCfg(id, true)
      .do(
        (test) =>  { this.counterItems++},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
    this.reloadData();
  }

  saveDeviceCfg() {
    if (this.deviceForm.valid) {
      this.deviceCfgService.addDeviceCfg(this.deviceForm.value)
        .subscribe(data => { console.log(data) },
        err => {
          console.log(err);
        },
        () => { this.editmode = "list"; this.reloadData() }
        );
    }
  }

  updateAllSelectedItems(mySelectedArray,field,value, append?) {
    let obsArray = [];
    this.counterItems = 0;
    this.isRequesting = true;
    if (!append)
    for (let component of mySelectedArray) {
      component[field] = value;
      obsArray.push(this.updateDeviceCfg(true,component));
    } else {
      let tmpArray = [];
      if(!Array.isArray(value)) value = value.split(',');
      console.log(value);
      for (let component of mySelectedArray) {
        console.log(value);
        //check if there is some new object to append
        let newEntries = _.differenceWith(value,component[field],_.isEqual);
        tmpArray = newEntries.concat(component[field])
        console.log(tmpArray);
        component[field] = tmpArray;
        obsArray.push(this.updateDeviceCfg(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateDeviceCfg(recursive?, component?) {
    if(!recursive) {
      if (this.deviceForm.valid) {
        var r = true;
        if (this.deviceForm.value.ID != this.oldID) {
          r = confirm("Changing Influx Server ID from " + this.oldID + " to " + this.deviceForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.deviceCfgService.editDeviceCfg(this.deviceForm.value, this.oldID, true)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.deviceCfgService.editDeviceCfg(component, component.ID,true)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
  }


  testDeviceCfgConnection() {
    this.deviceCfgService.testDeviceCfg(this.deviceForm.value, true)
    .subscribe(
    data =>  this.alertHandler = {msg: 'Influx Version: '+data['Message'], result : data['Result'], elapsed: data['Elapsed'], type: 'success', closable: true},
    err => {
        let error = err.json();
        this.alertHandler = {msg: error['Message'], elapsed: error['Elapsed'], result : error['Result'], type: 'danger', closable: true}
      },
    () =>  { console.log("DONE")}
  );

  }

  genericForkJoin(obsArray: any) {
    Observable.forkJoin(obsArray)
              .subscribe(
                data => {
                  this.selectedArray = [];
                  this.reloadData()
                },
                err => console.error(err),
              );
  }
  getInfluxServersforDevices() {
    this.influxServerService.getInfluxServer(null)
      .subscribe(
      data => {
      //  this.influxservers = data;
        this.selectinfluxservers = [];
        this.selectinfluxservers = this.createMultiselectArray(data,'ID','ID','Description');
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  createMultiselectArray(tempArray,id?,name?,description?) : any {
    let myarray = [];
    for (let entry of tempArray) {
      myarray.push({ 'id': id ? entry[id] : entry.ID, 'name': name? entry[name] : entry.ID, 'extraData': description ? entry[description] : '' });
    }
    return myarray;
  }


}
