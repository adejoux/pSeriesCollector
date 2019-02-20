import { FormBuilder, Validators, FormArray, FormGroup, FormControl } from '@angular/forms';
import { ValidationService } from './custom-validation/validation.service'

export class AvailableTableActions {

  //AvailableOptions result depeding on component type
  public availableOptions: Array<any>;

  // type can be : device,...
  // data is the passed extraData when declaring AvailableTableActions on each component
  checkComponentType(type, data?): any {
    switch (type) {
      case 'influxcfg':
        return this.getInfluxAvailableActions();
      case 'hmcservercfg':
        return this.getHMCServerAvailableActions(data);
      case 'devicecfg':
        return this.getDeviceAvailableActions(data);
      default:
        return null;
    }
  }

  constructor(componentType: string, extraData?: any) {
    console.log(extraData);
    this.availableOptions = this.checkComponentType(componentType, extraData);
  }

  getInfluxAvailableActions(data?: any): any {
    let tableAvailableActions = [
      //Remove Action
      {
        'title': 'Remove', 'content':
          { 'type': 'button', 'action': 'RemoveAllSelected' }
      },
      //Change Property Action
      {
        'title': 'Change property', 'content':
          {
            'type': 'selector', 'action': 'ChangeProperty', 'options': [
              {
                'title': 'Precision', 'type': 'boolean', 'options': [
                  'h', 'm', 's', 'ms', 'u', 'ns']
              },
              {
                'title': 'Retention', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.required)
                  })
              },
              {
                'title': 'Timeout', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator]))
                  })
              }
            ]
          },
      }
    ];
    return tableAvailableActions;
  }

  getHMCServerAvailableActions(data?: any): any {
    let tableAvailableActions = [
      //Remove Action
      {
        'title': 'Remove', 'content':
          { 'type': 'button', 'action': 'RemoveAllSelected' }
      },
      //Change Property Action
      {
        'title': 'Change property', 'content':
          {
            'type': 'selector', 'action': 'ChangeProperty', 'options': [
              {
                'title': 'Active', 'type': 'boolean', 'options': [
                  'true', 'false'
                ]
              },
              {
                'title': 'LogLevel', 'type': 'boolean', 'options': [
                  'panic', 'fatal', 'error', 'warning', 'info', 'debug'
                ]
              },
              {
                'title': 'ManagedSystemsOnly', 'type': 'boolean', 'options': [
                  'true', 'false'
                ]
              },
              {
                'title': 'Freq', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator]))
                  })
              },
              {
                'title': 'UpdateScanFreq', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerNotZeroValidator]))
                  })
              },
              {
                'title': 'UpdateFltFreq', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([Validators.required, ValidationService.uintegerAndLessOneValidator]))
                  })
              },
              {
                'title': 'DeviceTagValue', 'type': 'boolean', 'options': [
                  'id', 'host'
                ]
              },
              {
                'title': 'DeviceTagName', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.required)
                  })
              },
              {
                'title' : 'OutDB', 'type':'single-multiselector', 'options' :
                data
              },
              {
                'title': 'User', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.required)
                  })
              },
              {
                'title': 'Password', 'type': 'input-password', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.required)
                  })
              },
              {
                'title': 'ExtraTags', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
                  })
              }
            ]
          },
      },
      //AppendProperty
      {
        'title': 'AppendProperty', 'content':
          {
            'type': 'selector', 'action': 'AppendProperty', 'options': [
             {
                'title': 'ExtraTags', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
                  })
              }
            ]
          }
      }
    ];
    return tableAvailableActions;
  }


  getDeviceAvailableActions(data?: any): any {
    let tableAvailableActions = [
      //Remove Action
      {
        'title': 'Remove', 'content':
          { 'type': 'button', 'action': 'RemoveAllSelected' }
      },
      //Change Property Action
      {
        'title': 'Change property', 'content':
          {
            'type': 'selector', 'action': 'ChangeProperty', 'options': [
              {
                'title': 'EnableHMCStats', 'type': 'boolean', 'options': [
                  'true', 'false'
                ]
              },
              {
                'title': 'EnableNmonStats', 'type': 'boolean', 'options': [
                  'true', 'false'
                ]
              },
              {
                'title': 'NmonLogLevel', 'type': 'boolean', 'options': [
                  'panic', 'fatal', 'error', 'warning', 'info', 'debug'
                ]
              },
              {
                'title': 'NmonProtDebug', 'type': 'boolean', 'options': [
                  'panic', 'fatal', 'error', 'warning', 'info', 'debug'
                ]
              },
              {
                'title': 'NmonFilePath', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('')
                  })
              },
              {
                'title': 'NmonFreq', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('',ValidationService.uintegerNotZeroValidator)
                  })
              },
              {
                'title' : 'NmonOutDB', 'type':'single-multiselector', 'options' :
                data
              },
              {
                'title': 'NmonSSHUser', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('')
                  })
              },
              {
                'title': 'NmonSSHKey', 'type': 'input-password', 'options':
                  new FormGroup({
                    formControl: new FormControl('')
                  })
              },
              {
                'title': 'ExtraTags', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
                  })
              },
              {
                'title': 'NmonFilters', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([ValidationService.noWhiteSpaces,ValidationService.isValidRegexArray]))
                  })
              }
            ]
          },
      },
      //AppendProperty
      {
        'title': 'AppendProperty', 'content':
          {
            'type': 'selector', 'action': 'AppendProperty', 'options': [
             {
                'title': 'ExtraTags', 'type': 'input', 'options':
                  new FormGroup({
                    formControl: new FormControl('', Validators.compose([ValidationService.noWhiteSpaces, ValidationService.extraTags]))
                  })
              }
            ]
          }
      }
    ];
    return tableAvailableActions;
  }
}
