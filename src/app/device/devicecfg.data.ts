export const DeviceCfgComponentConfig: any =
  {
    'name' : 'Devices',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Dev  Name', name: 'Name' },
      { title: 'Serial Number', name: 'SerialNumber' },
      { title: 'OS Version',name:'OSVersion'},
      { title: 'type', name: 'Type' },
      { title: 'Location', name: 'Location' },
      { title: 'Enable HMC stats', name: 'EnableHMCStats' },
      { title: 'Enable Nmon Stats', name: 'EnableNmonStats' },
      { title: 'ExtraTags', name: 'ExtraTags' }
    ],
    'slug' : 'devicecfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]