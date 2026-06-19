import { Module, Global, OnApplicationShutdown } from '@nestjs/common';
import { KafkaService } from './kafka.service';

@Global()
@Module({
  providers: [KafkaService],
  exports: [KafkaService],
})
export class KafkaModule implements OnApplicationShutdown {
  constructor(private readonly kafka: KafkaService) {}

  async onApplicationShutdown(): Promise<void> {
    await this.kafka.disconnect();
  }
}
