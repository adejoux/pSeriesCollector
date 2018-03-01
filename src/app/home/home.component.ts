import { Component, ViewChild,ViewContainerRef } from '@angular/core';
import { NgSwitch, NgSwitchCase, NgSwitchDefault } from '@angular/common';
import { Router } from '@angular/router';
import { Observable } from 'rxjs/Observable';
import { BlockUIService } from '../common/blockui/blockui-service';
import { BlockUIComponent } from '../common/blockui/blockui-component';
import { ImportFileModal } from '../common/dataservice/import-file-modal';
import { ExportFileModal } from '../common/dataservice/export-file-modal';
import { HomeService } from './home.service';
import { AboutModal } from './about-modal'
import { WindowRef } from '../common/windowref';

//Menu Components  to load them dynamically
import { HMCServerComponent } from '../hmcserver/hmcserver.component';
import { InfluxServerCfgComponent } from '../influxserver/influxservercfg.component';
import { RuntimeComponent } from '../runtime/runtime.component';
import { DeviceCfgComponent } from '../device/devicecfg.component';
import { NavbarComponent } from './navbar/navbar.component';
import { SideMenuComponent } from './sidemenu/sidemenu.component';

declare var _:any;

@Component({
  selector: 'home',
  templateUrl: './home.component.html',
  styleUrls: [ './home.component.css' ],
  providers: [BlockUIService, HomeService]
})

export class HomeComponent {

  @ViewChild('blocker', { read: ViewContainerRef }) container: ViewContainerRef;
  @ViewChild('importFileModal') public importFileModal : ImportFileModal;
  @ViewChild('exportBulkFileModal') public exportBulkFileModal : ExportFileModal;

  @ViewChild('aboutModal') public aboutModal : AboutModal;
  @ViewChild('RuntimeComponent') public rt : any;

  nativeWindow: any
  response: string;
  item_type: string;
  version: RInfo;
  menuItems : Array<any> = [
  {'groupName' : 'Runtime', 'icon': 'glyphicon glyphicon-play', 'expanded': true, 'items':
    [
      {'title': 'Agent status', 'selector' : 'runtime-component', 'type' : 'component', 'data': RuntimeComponent}
    ]
  },
  {'groupName' : 'Server Config', 'icon': 'glyphicon glyphicon-cog', 'expanded': true, 'items':
  [
    {'title': 'Influx DB Servers ', 'selector' : 'ifxserver-component', 'type': 'component', 'data': InfluxServerCfgComponent},
    {'title': 'HMC Servers', 'selector' : 'hmcserver-component', 'type': 'component', 'data': HMCServerComponent},
    {'title': 'Devices', 'selector' : 'devicecfg-component', 'type': 'component', 'data': DeviceCfgComponent},
  ]
  },
  {'groupName' : 'Data Service', 'icon': 'glyphicon glyphicon-paste', 'expanded': true, 'items':
  [
    {'title': 'Export Data ', 'selector' : 'ifxserver-component', 'type': 'button', 'data': 'exportdata'},
    {'title': 'Import Data', 'selector' : 'hmcserver-component', 'type': 'button', 'data': 'importdata'},
  ]
  },
  {'groupName' : 'Actions', 'icon': 'glyphicon glyphicon-refresh', 'expanded': true, 'items':
   [
     {'title': 'Reload Config', 'type': 'button', 'data': 'reload'}
   ]
  }];


  mode : boolean = false;
  userIn : boolean = false;
  componentList = null;
  elapsedReload: string = '';
  lastReload: Date;

  constructor(private winRef: WindowRef,public router: Router, public _blocker: BlockUIService, public homeService: HomeService) {
    this.nativeWindow = winRef.nativeWindow;
    this.getFooterInfo();
    this.componentList = RuntimeComponent;
    this.item_type= "runtime-component";
  }

  link(url: string) {
    this.nativeWindow.open(url);
  }

  expandMenu(i : any) : boolean{
    this.menuItems[i].expanded = !this.menuItems[i].expanded;
    return this.menuItems[i].expanded;
  }

  logout() {
    this.homeService.userLogout()
    .subscribe(
    response => {
      this.router.navigate(['/sign-in']);
    },
    error => {
      alert(error.text());
      console.log(error.text());
    }
    );
  }
  changeModeMenu() {
    this.mode = !this.mode
  }

  clickMenu(menuItem : any) : void {
    this.componentList = menuItem.data;
  }

  clickButton(menuItem : any) : void {
    switch (menuItem.data) {
      case 'reload':
        this.reloadConfig();
      break;
      case 'importdata':
        this.showImportModal();
      break;
      case 'exportdata':
        this.showExportBulkModal();
      break;
    }
  }

  showImportModal() {
    this.importFileModal.initImport();
  }

  showExportBulkModal() {
    this.exportBulkFileModal.initExportModal(null, false);
  }

  showAboutModal() {
    this.aboutModal.showModal(this.version);
  }

  reloadConfig() {
    this._blocker.start(this.container, "Reloading Conf. Please wait...");
    if (this.rt) this.rt.updateRuntimeInfo(null,null,false);
    this.homeService.reloadConfig()
    .subscribe(
    response => {
      this.lastReload = new Date();
      this.elapsedReload = response;
      this._blocker.stop();
    },
    error => {
      this._blocker.stop();
      alert(error.text());
      console.log(error.text());
    }
    );
  }

  getFooterInfo() {
    this.homeService.getInfo()
    .subscribe(data => {
      this.version = data;
      this.userIn = true;
    },
    err => console.error(err),
    () =>  {}
    );
  }
}
