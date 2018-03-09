export const HMCServerComponentConfig: any =
  {
    'name' : 'HMCServer',
    'table-columns' : [
      { 'title': 'ID', 'name': 'ID' },
      { 'title': 'Host', 'name': 'Host' },
      { 'title': 'Port', 'name': 'Port' },
      { 'title': 'Active', 'name': 'Active' },
      { 'title': 'Systems Only', 'name': 'ManagedSystemsOnly' },
      { 'title': 'User', 'name': 'User' },
      { 'title': 'Freq', 'name': 'Freq' },
      { 'title': 'Update Scan', 'name': 'UpdateScanFreq' },
      { 'title': 'OutDB', 'name': 'OutDB' },
      { 'title': 'Extra Tags', 'name': 'ExtraTags' }
    ],
    'slug' : 'hmcservercfg'
  };
  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'},
    {'name':'importHMCDevices', 'type':'icon', 'icon' : 'glyphicon glyphicon-import text-info', 'tooltip': 'Import Devices'},
  ]