export const RuntimeComponentConfig: any =
  {
    'name' : 'Runtime',
    'table-columns' : [
      { title: 'ID', name: 'ID' },
      { title: 'TagMap', name: 'TagMap', tooltip: 'Num Measurements configured' },
      { title: 'Type', name: 'Type', tooltip: 'Type of device polling method' },
      { title: 'Metr.Sent', name: 'Counter0', tooltip: 'MetricSent all values had been sent' },
      { title: 'Metr.Errs', name: 'Counter1', tooltip: 'Metric Errors: number of metrics (taken as fields) with errors for all measurements' },
      { title: 'Meas.Errs', name: 'Counter3', tooltip: 'MeasurementSentErrors: number of measuremenets  formatted with errors ' },
      { title: 'G.Time', name: 'Counter5', tooltip: 'CycleGatherDuration time: elapsed time taken to get all measurement info', transform: 'elapsedseconds' },
    ],
  }; 

  export const TableRole : string = 'runtime';

  export const ExtraActions: Array<any> = [
    { title: 'SetActive', type: 'boolean', content: { enabled: '<i class="glyphicon glyphicon-pause"></i>', disabled: '<i class="glyphicon glyphicon-play"></i>' }, 'property': 'DeviceActive' }
    /*{ title: 'SnmpReset', type: 'button', content: { enabled: 'Reset' } }*/
  ];
 
  export const CounterDef: CounterType[] = [
    /*0*/    { show: true, id: "MetricSent", label: "Metric Sent", type: "counter", tooltip: "number of metrics sent (taken as fields) for all measurements" },
    /*1*/    { show: true, id: "MetricSentErrors", label: "Metric Sent Errors", type: "counter", tooltip: "number of metrics  (taken as fields) with errors forall measurements" },
    /*2*/    { show: true, id: "MeasurementSent", label: "Measurement sent", type: "counter", tooltip: "(number of  measurements build to send as a sigle request sent to the backend)" },
    /*3*/    { show: true, id: "MeasurementSentErrors", label: "Measurement sent Errors", type: "counter", tooltip: "(number of measuremenets  formatted with errors )" },
    /*4*/    { show: false, id: "CycleGatherStartTime", label: "Cycle Gather Start Time", type: "time", tooltip: "Last gather time " },
    /*5*/    { show: true, id: "CycleGatherDuration", label: "Cycle Gather Duration", type: "duration", tooltip: "elapsed time taken to get all measurement info" },
    /*6*/    { show: false, id: "BackEndSentStartTime", label: "BackEnd (influxdb) Sent Start Time", type: "time", tooltip: "Last sent time" },
    /*7*/    { show: true, id: "BackEndSentDuration", label: "BackEnd (influxdb) Sent Duration", type: "duration", tooltip: "elapsed time taken to send data to the db backend" },
    /*8*/    { show: false, id: "ScanStartTime", label: "Device Scan Start Time", type: "time", tooltip: "Last scan time" },
    /*9*/    { show: true, id: "ScanDuration", label: "Device Scan Duration", type: "duration", tooltip: "elapsed time taken to scan devices" },
   ];