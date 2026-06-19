import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { Kafka, Producer } from 'kafkajs';
import { v4 as uuidv4 } from 'uuid';

const TOPIC_USER_REGISTERED = 'crypto2krw.core.user.registered';

interface CloudEvent<T> {
  specversion: '1.0';
  id: string;
  source: string;
  type: string;
  time: string;
  datacontenttype: 'application/json';
  data: T;
}

@Injectable()
export class KafkaService implements OnModuleInit {
  private readonly logger = new Logger(KafkaService.name);
  private producer: Producer;

  constructor() {
    const kafka = new Kafka({
      clientId: process.env.KAFKA_CLIENT_ID ?? 'auth-service',
      brokers: (process.env.KAFKA_BROKERS ?? 'localhost:9093').split(','),
      retry: { retries: 5 },
    });
    this.producer = kafka.producer({
      idempotent: true,  // 프로듀서 멱등성 활성화
    });
  }

  async onModuleInit(): Promise<void> {
    await this.producer.connect();
    this.logger.log('Kafka producer connected');
  }

  async disconnect(): Promise<void> {
    await this.producer.disconnect();
  }

  async publishUserRegistered(userId: string, email: string): Promise<void> {
    const event: CloudEvent<{ user_id: string; email: string }> = {
      specversion: '1.0',
      id: uuidv4(),
      source: 'channel/auth-service',
      type: 'crypto2krw.core.user.registered',
      time: new Date().toISOString(),
      datacontenttype: 'application/json',
      data: { user_id: userId, email },
    };

    await this.producer.send({
      topic: TOPIC_USER_REGISTERED,
      messages: [{ key: userId, value: JSON.stringify(event) }],
    });

    this.logger.log(`published user.registered for user=${userId}`);
  }
}
