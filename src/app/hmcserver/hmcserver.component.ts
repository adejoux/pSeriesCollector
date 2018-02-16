import { Component, ChangeDetectionStrategy, ViewChild, OnInit } from '@angular/core';
import { FormBuilder, Validators} from '@angular/forms';
import { FormArray, FormGroup, FormControl} from '@angular/forms';
import { IMultiSelectOption, IMultiSelectSettings, IMultiSelectTexts } from '../common/multiselect-dropdown';

import { HMCServerService } from './hmcserver.service';
import { ValidationService } from '../common/custom-validation/validation.service'
import { ExportServiceCfg } from '../common/dataservice/export.service'

import { GenericModal } from '../common/custom-modal/generic-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { Observable } from 'rxjs/Rx';

import { TableListComponent } from '../common/table-list.component';
import { HMCServerComponentConfig, TableRole, OverrideRoleActions } from './hmcserver.data';
import { InfluxServerService } from '../influxserver/influxservercfg.service';

declare var _:any;

@Component({
  selector: 'hmcserver-component',
  providers: [HMCServerService, ValidationService,InfluxServerService],
  templateUrl: './hmcserver.component.html',
  styleUrls: ['../../css/component-styles.css']
})

export class HMCServerComponent implements OnInit {
  @ViewChild('viewModal') public viewModal: GenericModal;
  @ViewChild('viewModalDelete') public viewModalDelete: GenericModal;
  @ViewChild('exportFileModal') public exportFileModal : ExportFileModal;

  public alertHandler:any;
  public editmode: string; //list , create, modify
  public componentList: Array<any>;
  public filter: string;
  public sampleComponentForm: any;
  public counterItems : number = null;
  public counterErrors: any = [];
  public defaultConfig : any = HMCServerComponentConfig;
  public tableRole : any = TableRole;
  public overrideRoleActions: any = OverrideRoleActions;
  public selectedArray : any = [];
  public selectinfluxservers: IMultiSelectOption[] = [];
  private mySettingsInflux: IMultiSelectSettings = {
    singleSelect: true,
  };

  public data : Array<any>;
  public isRequesting : boolean;

  private builder;
  private oldID : string;

  ngOnInit() {
    this.editmode = 'list';
    this.reloadData();
  }

  constructor(public influxServerService: InfluxServerService, public hmcserverService: HMCServerService, public exportServiceCfg : ExportServiceCfg, builder: FormBuilder) {
    this.builder = builder;
  }

  createStaticForm() {
    this.sampleComponentForm = this.builder.group({
      ID: [this.sampleComponentForm ? this.sampleComponentForm.value.ID : '', Validators.required],
      Host: [this.sampleComponentForm ? this.sampleComponentForm.value.Host : '', Validators.required],
      Port: [this.sampleComponentForm ? this.sampleComponentForm.value.Port : 12443, Validators.required],
      User: [this.sampleComponentForm ? this.sampleComponentForm.value.User : 'hmcuser', Validators.required],
      Password: [this.sampleComponentForm ? this.sampleComponentForm.value.Password : '', Validators.required],
      Active: [this.sampleComponentForm ? this.sampleComponentForm.value.Active : 'true', Validators.required],
      ManagedSystemsOnly: [this.sampleComponentForm ? this.sampleComponentForm.value.ManagedSystemsOnly : 'false', Validators.required],
      Freq: [this.sampleComponentForm ? this.sampleComponentForm.value.Freq : 60, Validators.required],
      OutDB: [this.sampleComponentForm ? this.sampleComponentForm.value.OutDB : '', Validators.required],
      LogLevel: [this.sampleComponentForm ? this.sampleComponentForm.value.LogLevel : 'info', Validators.required],
      HMCAPIDebug: [this.sampleComponentForm ? this.sampleComponentForm.value.HMCAPIDebug : 'false', Validators.required],
      DeviceTagName: [this.sampleComponentForm ? this.sampleComponentForm.value.DeviceTagName : '', Validators.required],
      DeviceTagValue: [this.sampleComponentForm ? this.sampleComponentForm.value.DeviceTagValue : 'id'],
      ExtraTags: [this.sampleComponentForm ? (this.sampleComponentForm.value.ExtraTags ? this.sampleComponentForm.value.ExtraTags : "" ) : "" , Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags])],
      Description: [this.sampleComponentForm ? this.sampleComponentForm.value.Description : '']
    });
  }

  testHMCServerConnection(data:any) {
    this.hmcserverService.testHMCServer(data, true)
    .subscribe(
    data =>  this.alertHandler = {msg: 'HCM Version: '+data['Message'], result : data['Result'], elapsed: data['Elapsed'], type: 'success', closable: true},
    err => {
        let error = err.json();
        this.alertHandler = {msg: error['Message'], elapsed: error['Elapsed'], result : error['Result'], type: 'danger', closable: true}
      },
    () =>  { console.log("DONE")}
  );
}

  reloadData() {
    // now it's a simple subscription to the observable
  this.hmcserverService.getHMCServerItem(null)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.componentList = data
        this.data = data;
        this.editmode = "list";
      },
      err => console.error(err),
      () => console.log('DONE')
      );
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
        this.editSampleItem(action.event);
      break;
      case 'remove':
        this.removeItem(action.event);
      break;
      case 'tableaction':
        this.applyAction(action.event, action.data);
      break;
    }
  }


  applyAction(action : any, data? : Array<any>) : void {
    this.selectedArray = data || [];
    switch(action.action) {
       case "RemoveAllSelected": {
          this.removeAllSelectedItems(this.selectedArray);
          break;
       }
       case "ChangeProperty": {
          this.updateAllSelectedItems(this.selectedArray,action.field,action.value)
          break;
       }
       case "AppendProperty": {
         this.updateAllSelectedItems(this.selectedArray,action.field,action.value,true);
       }
       default: {
          break;
       }
    }
  }

  viewItem(id) {
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
      this.deleteSampleItem(myArray[i].ID,true);
      obsArray.push(this.deleteSampleItem(myArray[i].ID,true));
    }
    this.genericForkJoin(obsArray);
  }

  removeItem(row) {
    let id = row.ID;
    console.log('remove', id);
    this.hmcserverService.checkOnDeleteHMCServerItem(id)
      .subscribe(
        data => {
        this.viewModalDelete.parseObject(data)
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

  editSampleItem(row) {
    let id = row.ID;
    this.getInfluxServersforDevices()
    this.hmcserverService.getHMCServerItemById(id)
      .subscribe(data => {
        this.sampleComponentForm = {};
        this.sampleComponentForm.value = data;
        this.oldID = data.ID
        this.createStaticForm();
        this.editmode = "modify";
      },
      err => console.error(err)
      );
 	}

  deleteSampleItem(id, recursive?) {
    if (!recursive) {
    this.hmcserverService.deleteHMCServerItem(id)
      .subscribe(data => { },
      err => console.error(err),
      () => { this.viewModalDelete.hide(); this.reloadData() }
      );
    } else {
      return this.hmcserverService.deleteHMCServerItem(id)
      .do(
        (test) =>  { this.counterItems++; console.log(this.counterItems)},
        (err) => { this.counterErrors.push({'ID': id, 'error' : err})}
      );
    }
  }

  cancelEdit() {
    this.editmode = "list";
    this.reloadData();
  }

  saveSampleItem() {
    if (this.sampleComponentForm.valid) {
      this.hmcserverService.addHMCServerItem(this.sampleComponentForm.value)
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
      obsArray.push(this.updateSampleItem(true,component));
    } else {
      let tmpArray = [];
      if(!Array.isArray(value)) value = value.split(',');
      for (let component of mySelectedArray) {
        //check if there is some new object to append
        let newEntries = _.differenceWith(value,component[field],_.isEqual);
        tmpArray = newEntries.concat(component[field])
        component[field] = tmpArray;
        obsArray.push(this.updateSampleItem(true,component));
      }
    }
    this.genericForkJoin(obsArray);
    //Make sync calls and wait the result
    this.counterErrors = [];
  }

  updateSampleItem(recursive?, component?) {
    if(!recursive) {
      if (this.sampleComponentForm.valid) {
        var r = true;
        if (this.sampleComponentForm.value.ID != this.oldID) {
          r = confirm("Changing HMCServer Instance ID from " + this.oldID + " to " + this.sampleComponentForm.value.ID + ". Proceed?");
        }
        if (r == true) {
          this.hmcserverService.editHMCServerItem(this.sampleComponentForm.value, this.oldID)
            .subscribe(data => { console.log(data) },
            err => console.error(err),
            () => { this.editmode = "list"; this.reloadData() }
            );
        }
      }
    } else {
      return this.hmcserverService.editHMCServerItem(component, component.ID)
      .do(
        (test) =>  { this.counterItems++ },
        (err) => { this.counterErrors.push({'ID': component['ID'], 'error' : err['_body']})}
      )
      .catch((err) => {
        return Observable.of({'ID': component.ID , 'error': err['_body']})
      })
    }
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
