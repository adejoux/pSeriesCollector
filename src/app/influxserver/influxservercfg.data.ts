export const InfluxServerCfgComponentConfig: any =
  {
    'name' : 'Influx Server',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'Host', name: 'Host' },
      { title: 'Port', name: 'Port' },
      { title: 'Enable SSL',name:'EnableSSL'},
      { title: 'DB', name: 'DB' },
      { title: 'User', name: 'User' },
      { title: 'Retention', name: 'Retention' },
      { title: 'Precision', name: 'Precision' },
      { title: 'Timeout', name: 'Timeout' },
      { title: 'User Agent', name: 'UserAgent' }
    ],
    'slug' : 'influxcfg'
  }; 

  export const TableRole : string = 'fulledit';
  export const OverrideRoleActions : Array<Object> = [
    {'name':'export', 'type':'icon', 'icon' : 'glyphicon glyphicon-download-alt text-default', 'tooltip': 'Export item'},
    {'name':'view', 'type':'icon', 'icon' : 'glyphicon glyphicon-eye-open text-success', 'tooltip': 'View item'},
    {'name':'edit', 'type':'icon', 'icon' : 'glyphicon glyphicon-edit text-warning', 'tooltip': 'Edit item'},
    {'name':'remove', 'type':'icon', 'icon' : 'glyphicon glyphicon glyphicon-remove text-danger', 'tooltip': 'Remove item'}
  ]