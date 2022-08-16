import { NgModule } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { JobDetailsRoutingModule } from './job-details-routing'
import { CommonModule } from '@angular/common';
import { JobDetailComponent } from './job-detail.component'
import { ShardModule } from 'src/app/shared/shard/shard.module'
@NgModule({
  declarations: [
    JobDetailComponent
  ],
  imports: [
    CommonModule,
    JobDetailsRoutingModule,
    ClarityModule,
    FormsModule,
    ShardModule
  ],
})
export class JobDetailModule {}
