
import { ChangeDetectionStrategy, Component, ViewChild, ChangeDetectorRef, OnDestroy } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { RuntimeService } from './runtime.service';
import { ItemsPerPageOptions } from '../common/global-constants';

import { SpinnerComponent } from '../common/spinner';
import { HMCServerService } from '../hmcserver/hmcserver.service';

import { RuntimeComponentConfig, TableRole, ExtraActions, CounterDef } from './runtime.data';

declare var _: any;
@Component({
  selector: 'runtime',
  providers: [RuntimeService, HMCServerService],
  templateUrl: './runtimeview.html',
  styleUrls: ['./runtimeeditor.css'],
})


export class RuntimeComponent implements OnDestroy {
  
  itemsPerPageOptions: any = ItemsPerPageOptions;
  public isRefreshing: boolean = true;

  public selected : Object = {'page' : 1, 'itemsPerPage' : 100 };
  public selectedTab : string =  'sections';
  public defaultConfig : any = RuntimeComponentConfig;
  public tableRole : any = TableRole;
  public extraActions: any = ExtraActions;
  public oneAtATime: boolean = true;
  editmode: string; //list , create, modify
  isRequesting: boolean = false;
  runtime_devs: Array<any>;

  mySubscription: any;
  filter: string;
  measActive: number = 0;
  runtime_dev: any;
  subItem: any;
  islogLevelChanged: boolean = false;
  newLogLevel: string = null;
  loglLevelArray: Array<string> = [
    'panic',
    'fatal',
    'error',
    'warning',
    'info',
    'debug'
  ];
  maxrep: any = '';
  counterDef = CounterDef;

  //TABLE
  private data: Array<any> = [];
  public activeDevices: number;
  public noConnectedDevices: number;
  public dataTable: Array<any> = [];
  public finalData: Array<Array<any>> = [];
  public columns: Array<any> = [];
  public tmpcolumns: Array<any> = [];

  public refreshRuntime: any = {
    'Running': false,
    'LastUpdate': new Date()
  }
  public intervalStatus: any

  public rows: Array<any> = [];
  public page: number = 1;
  public itemsPerPage: number = 20;
  public maxSize: number;
  public numPages: number = 1;
  public length: number = 0;
  public myFilterValue: any;
  public activeFilter: boolean = false;
  public deactiveFilter: boolean = false;
  public noConnectedFilter: boolean = false;

  //Set config
  public config: any = {
    paging: true,
    sorting: { columns: this.columns },
    filtering: { filterString: '' },
    className: ['table-striped', 'table-bordered']
  };

  constructor(public runtimeService: RuntimeService, builder: FormBuilder, private ref: ChangeDetectorRef, public hmcServerService: HMCServerService) {
    this.editmode = 'list';
    this.reloadData();
  }


  public changePage(page: any, data: Array<any> = this.data): Array<any> {
    //Check if we have to change the actual page
    let maxPage = Math.ceil(data.length / this.itemsPerPage);
    if (page.page > maxPage && page.page != 1) this.page = page.page = maxPage;

    let start = (page.page - 1) * page.itemsPerPage;
    let end = page.itemsPerPage > -1 ? (start + page.itemsPerPage) : data.length;
    return data.slice(start, end);
  }

  public changeSort(data: any, config: any): any {
    if (!config.sorting) {
      return data;
    }
    let columns = this.config.sorting.columns || [];
    let columnName: string = void 0;
    let sort: string = void 0;

    for (let i = 0; i < columns.length; i++) {
      if (columns[i].sort !== '' && columns[i].sort !== false) {
        columnName = columns[i].name;
        sort = columns[i].sort;
      }
    }

    if (!columnName) {
      return data;
    }

    // simple sorting
    return data.sort((previous: any, current: any) => {
      if (previous[columnName] > current[columnName]) {
        return sort === 'desc' ? -1 : 1;
      } else if (previous[columnName] < current[columnName]) {
        return sort === 'asc' ? -1 : 1;
      }
      return 0;
    });
  }

  public changeFilter(data: any, config: any): any {
    let filteredData: Array<any> = data;
    this.columns.forEach((column: any) => {
      if (column.filtering) {
        filteredData = filteredData.filter((item: any) => {
          return item[column.name].match(column.filtering.filterString);
        });
      }
    });

    if (!config.filtering) {
      return filteredData;
    }

    if (config.filtering.columnName) {
      return filteredData.filter((item: any) =>
        item[config.filtering.columnName].match(this.config.filtering.filterString));
    }

    let tempArray: Array<any> = [];
    filteredData.forEach((item: any) => {
      let flag = false;
      this.columns.forEach((column: any) => {
        if (item[column.name] === null) {
          item[column.name] = '--'
        }
        if (item[column.name].toString().match(this.config.filtering.filterString)) {
          flag = true;
        }
      });
      if (flag) {
        tempArray.push(item);
      }
    });
    filteredData = tempArray;
    return filteredData;
  }


  resetTabs(tab : string) {
    this.selected = {'page':1, 'itemsPerPage': 100}
    this.selectedTab = tab;
  }

  changeItemsPerPage(items) {
    if (items) this.itemsPerPage = parseInt(items);
    else this.itemsPerPage = this.length;
    let maxPage = Math.ceil(this.length / this.itemsPerPage);
    if (this.page > maxPage) this.page = maxPage;
    this.onChangeTable(this.config);
  }

  public onChangeTable(config: any, page: any = { page: this.page, itemsPerPage: this.itemsPerPage }): any {
    if (config.filtering) {
      Object.assign(this.config.filtering, config.filtering);
    }
    if (config.sorting) {
      Object.assign(this.config.sorting, config.sorting);
    }
    let filteredData = this.changeFilter(this.data, this.config);
    let sortedData = this.changeSort(filteredData, this.config);
    this.rows = page && this.config.paging ? this.changePage(page, sortedData) : sortedData;
    this.length = sortedData.length;
    this.activeDevices = sortedData.filter((item) => { return item.DeviceActive }).length
    this.noConnectedDevices = sortedData.filter((item) => { if (item.DeviceActive === true && item.DeviceConnected === false) return true }).length
  }

  public onExtraActionClicked(data: any) {
    switch (data.action) {
      case 'SetActive':
        this.changeActiveDevice(data.row.ID, !data.row.DeviceActive)
        break;
      default:
        break;
    }
    console.log(data);
  }

  onResetFilter(): void {
    this.page = 1;
    this.myFilterValue = "";
    this.config.filtering = { filtering: { filterString: '' } };
    this.onChangeTable(this.config);
  }

  initRuntimeInfo(id: string, meas: number, isRequesting?: boolean) {
    //Reset params
    this.editmode = 'view';
    if (isRequesting) {
      this.isRequesting = isRequesting;
      this.runtime_dev = null;
    }
    this.isRefreshing = false;
    this.refreshRuntime.Running = false;
    this.selectedTab = "sections";
    clearInterval(this.intervalStatus);
    if (!this.mySubscription) {
      this.loadRuntimeById(id);
    }
  }

  updateRuntimeInfo(id: string, status: boolean) {
    clearInterval(this.intervalStatus);
    this.refreshRuntime.Running = status;
    if (this.refreshRuntime.Running) {
      this.isRefreshing = true;
      this.refreshRuntime.LastUpdate = new Date();
      //Cargamos interval y dejamos actualizando la información:
      this.intervalStatus = setInterval(() => {
        this.isRefreshing = false;
        setTimeout(() => {
          this.isRefreshing = true;
        }, 2000);
        this.refreshRuntime.LastUpdate = new Date();
        this.loadRuntimeById(id);
        this.ref.markForCheck();
      }, Math.max(5000, this.runtime_dev['Freq'] * 1000)); //lowest update rate set to 5 sec
    } else {
      this.isRefreshing = false;
      clearInterval(this.intervalStatus);
    }
  }

  tableCellParser (data: any, type: string) {
    if (type === "MULTISTRINGPARSER") {
      var test: any = '<ul class="list-unstyled">';
      for (var i of data) {
          test +="<li>"
          test +="<span class=\"badge\">"+ ( i["IType"] === 'T' ? "Tag" : "Field" ) +"</span><b>"+ i["IName"] + " :</b>" +  i["Value"];
          test += "</li>";
      }
      test += "</ul>"
      return test
    }
    return ""
  }

  loadRuntimeById(id: string) {
    this.mySubscription = this.runtimeService.getRuntimeById(id)
      .subscribe(
      data => {
        this.isRequesting = false;
        this.mySubscription = null;
        this.finalData = [];
        this.runtime_dev = data;
        this.runtime_dev.ID = id;
        if (!this.refreshRuntime.Running) this.updateRuntimeInfo(id,true);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  changeActiveDevice(id, event) {
    console.log("ID,event", id, event);

    this.runtimeService.changeDeviceActive(id, event)
      .subscribe(
      data => {
        _.forEach(this.runtime_devs, function (d, key) {
          console.log(d, key)
          if (d.ID == id) {
            d.DeviceActive = !d.DeviceActive;
            return false
          }
        })
        console.log(this.runtime_devs);
        if (this.runtime_dev != null) {
          this.runtime_dev.DeviceActive = !this.runtime_dev.DeviceActive;
        }

      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  onChangeLogLevel(level) {
    this.islogLevelChanged = true;
    this.newLogLevel = level;
  }

  changeLogLevel(id) {
    console.log("ID,event");
    this.runtimeService.changeLogLevel(id, this.newLogLevel)
      .subscribe(
      data => {
        this.runtime_dev.CurLogLevel = this.newLogLevel;
        this.islogLevelChanged = false;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  downloadLogFile(id) {
    console.log("Download Log file from device", id);
    this.runtimeService.downloadLogFile(id)
      .subscribe(
      data => {
        saveAs(data, id + ".log")
        console.log("download done")
      },
      err => {
        console.error(err)
        console.log("Error downloading the file.")
      },
      () => console.log('Completed file download.')
      );
  }

  forceDevScan(id) {
    console.log("ID,event", id, event);
    this.runtimeService.forceDevScan(id)
      .subscribe(
      data => {
        console.log("Device Scan update done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  forceDevReset(id, mode) {
    console.log("ID,event", id, event);
    this.runtimeService.forceDevReset(id, mode)
      .subscribe(
      data => {
        console.log("reset done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  forceGatherData(id) {
    console.log("force gather data", id, event);
    this.runtimeService.forceGatherData(id)
      .subscribe(
      data => {
        console.log("forced gather done")
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }


  changeStateDebug(id, event) {
    console.log("ID,event", id, event);
    this.runtimeService.changeStateDebug(id, event)
      .subscribe(
      data => {
        this.runtime_dev.StateDebug = !this.runtime_dev.StateDebug;
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }


  reloadData() {
    this.itemsPerPage = 20;
    this.isRequesting = true;
    if (this.mySubscription) {
      this.mySubscription.unsubscribe();
    }
    clearInterval(this.intervalStatus);
    this.editmode = "list"
    this.columns = this.defaultConfig['table-columns']
    this.filter = null;
    // now it's a simple subscription to the observable
    this.mySubscription = this.runtimeService.getRuntime(null)
      .subscribe(
      data => {
        this.mySubscription = null;
        this.runtime_devs = data
        this.data = this.runtime_devs;
        this.config.sorting.columns = this.columns,
          this.isRequesting = false;
        this.activeFilter = this.deactiveFilter = this.noConnectedFilter = false;
        this.onChangeTable(this.config);
      },
      err => console.error(err),
      () => console.log('DONE')
      );
  }

  toogleActiveFilter(option: string) {
    if (this.activeFilter === false && option === 'active') {
      this.noConnectedFilter = false;
      this.deactiveFilter = false;
      this.activeFilter = true;
      this.data = this.runtime_devs.filter((item) => { if (item.DeviceActive === true) return true })
    } else if (this.deactiveFilter === false && option === 'deactive') {
      this.noConnectedFilter = false;
      this.activeFilter = false;
      this.deactiveFilter = true;
      this.data = this.runtime_devs.filter((item) => { if (item.DeviceActive === false) return true })
    } else if (this.noConnectedFilter === false && option === 'noconnected') {
      this.noConnectedFilter = true;
      this.activeFilter = false;
      this.deactiveFilter = false;
      this.data = this.runtime_devs.filter((item) => { if (item.DeviceConnected === false && item.DeviceActive === true) return true })
    } else {
      this.data = this.runtime_devs;
      this.noConnectedFilter = false;
      this.activeFilter = false;
      this.deactiveFilter = false;
    }
    this.onChangeTable(this.config);
  }

  ngOnDestroy() {
    clearInterval(this.intervalStatus);
    if (this.mySubscription) this.mySubscription.unsubscribe();
  }

}
