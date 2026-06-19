import { Module } from '@nestjs/common';
import { DatabaseModule } from './database/database.module';
import { KafkaModule } from './kafka/kafka.module';
import { AuthModule } from './auth/auth.module';
import { UsersModule } from './users/users.module';
import { MerchantModule } from './merchant/merchant.module';
import { AdminModule } from './admin/admin.module';

@Module({
  imports: [
    DatabaseModule,
    KafkaModule,
    AuthModule,
    UsersModule,
    MerchantModule,
    AdminModule,
  ],
})
export class AppModule {}
