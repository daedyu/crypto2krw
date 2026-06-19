import { Module, Global } from '@nestjs/common';
import { Pool } from 'pg';

export const DB_POOL = 'DB_POOL';

@Global()
@Module({
  providers: [
    {
      provide: DB_POOL,
      useFactory: (): Pool => {
        const pool = new Pool({
          connectionString: process.env.DATABASE_URL,
          max: 10,
          idleTimeoutMillis: 30_000,
          connectionTimeoutMillis: 5_000,
        });

        pool.on('error', (err) => {
          console.error('pg pool error', err);
        });

        return pool;
      },
    },
  ],
  exports: [DB_POOL],
})
export class DatabaseModule {}
