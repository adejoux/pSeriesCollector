import { Component, Input, Output, Pipe, PipeTransform, ViewChild, EventEmitter } from '@angular/core';
import { FormGroup,FormControl } from '@angular/forms';
import { ModalDirective } from 'ngx-bootstrap';
import { WindowRef } from '../common/windowref';

@Component({
    selector: 'about-modal',
    template: `
      <div bsModal #childModal="bs-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myLargeModalLabel" aria-hidden="true">
          <div class="modal-dialog">
            <div class="modal-content">
              <div class="modal-header">
                <button type="button" class="close" (click)="childModal.hide()" aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title"><i class="glyphicon glyphicon-info-sign"></i> {{titleName}}</h4>
              </div>
              <div class="modal-body" *ngIf="info">
              <h4 class="text-primary"> <b>pseriescollector</b> </h4>
              <span> pSeriesCollector is a IBM PSeries Platform Metric collector system </span>
              <div class="text-right">
                <a href="javaScript:void(0);"  (click)="link('https://github.com/adejoux/pSeriesCollector')" class="text-link"> More info <i class="glyphicon glyphicon-plus-sign"></i></a>
              </div>
              <hr/>
              <h4> Release information </h4>
                <dl class="dl-horizontal">
                  <dt>Instance ID:</dt><dd>{{ info.InstanceID}}</dd>
                  <dt>Version:</dt><dd>{{info.Version}}</dd>
                  <dt>Commit:</dt><dd>{{info.Commit}}</dd>
                  <dt>Build Date:</dt><dd>{{ info.BuildStamp*1000 | date:'yyyy/M/d HH:mm:ss' }}</dd>
                  <dt>License:</dt><dd>MIT License</dd>
                </dl>
                <hr>
              <h4> Authors: </h4>
              <dl class="dl-horizontal">
                  <dt>Alain Djoux</dt>
                  <dd>
                    <a href="javascript:void(0);" (click)="link('https://github.com/adejoux')">GitHub</a>
                  </dd>
                  <dt>Toni Moreno</dt>
                  <dd>
                    <a href="javascript:void(0);" (click)="link('http://github.com/toni-moreno')">GitHub</a>
                  </dd>
                  <dt>Sergio Bengoechea</dt>
                  <dd>
                    <a href="javascript:void(0);" (click)="link('http://github.com/sbengo')">GitHub</a>
                  </dd>
              </dl>
              <hr>
              </div>
              <div class="modal-footer" *ngIf="showValidation === true">
               <button type="button" class="btn btn-primary" (click)="childModal.hide()">Close</button>
             </div>
            </div>
          </div>
        </div>`
})

export class AboutModal {
  @ViewChild('childModal') public childModal: ModalDirective;
  @Input() titleName : any;
  @Input() customMessage: string;
  @Input() showValidation: boolean;
  @Input() textValidation: string;

  @Output() public validationClicked:EventEmitter<any> = new EventEmitter();

  public info : RInfo;

  public validationClick(myId: string):void {
    this.validationClicked.emit(myId);
  }
  nativeWindow: any
  constructor(private winRef: WindowRef) {
    this.nativeWindow = winRef.nativeWindow;
  }

  showModal(info : RInfo){
    this.info = info;
    this.childModal.show();
  }

  link(url: string) {
    this.nativeWindow.open(url);
  }

  hide() {
    this.childModal.hide();
  }

}
