import { Component, Input, OnInit, ViewChild, ElementRef } from '@angular/core';
import { ClarityIcons, successStandardIcon, exclamationCircleIcon, exclamationTriangleIcon, infoCircleIcon } from '@cds/core/icon';
import { AppService } from '../../app.service'
ClarityIcons.addIcons(successStandardIcon);
ClarityIcons.addIcons(exclamationCircleIcon);
ClarityIcons.addIcons(exclamationTriangleIcon);
ClarityIcons.addIcons(infoCircleIcon);
@Component({
  selector: 'app-message',
  templateUrl: './message.component.html',
  styleUrls: ['./message.component.css']
})

//This component is for message pop up window
export class MessageComponent implements OnInit {

  constructor(public app: AppService) {
  }
  public classType: string[] = []
  public classCloseType: string[] = []
  public messageContent = ''
  ngOnInit(): void {
    this.classType = ['upc-message-' + this.messageType]
    this.classCloseType = ['close-' + this.messageType]
  }
  //message type: 'success',  'info',  'warning',  'hide', 'error'
  @Input() messageType: 'success' | 'info' | 'warning' | 'hide' | 'error' = 'info'
  @ViewChild('msg') msg!: ElementRef;
  public distory() {
    this.msg.nativeElement.style.display = 'none'
  }
}
