import { Module } from '@nestjs/common';
import { JwtModule } from '@nestjs/jwt';
import { MerchantController } from './merchant.controller';
import { MerchantService } from './merchant.service';

@Module({
  imports: [JwtModule.register({})],
  controllers: [MerchantController],
  providers: [MerchantService],
})
export class MerchantModule {}
