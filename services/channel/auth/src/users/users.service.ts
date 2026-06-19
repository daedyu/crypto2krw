import { Injectable, Inject, NotFoundException } from '@nestjs/common';
import { Pool } from 'pg';
import { DB_POOL } from '../database/database.module';

export interface UserProfile {
  id: string;
  email: string;
  kyc_status: string;
  status: string;
  created_at: string;
}

export interface WalletAddress {
  currency: string;
  address: string;
  network: string;
  payment_priority: number;
}

export interface BalanceEntry {
  currency: string;
  available_balance: string;
  locked_balance: string;
}

export interface TransactionRow {
  id: string;
  type: string;
  used_currency: string;
  used_amount: string;
  amount_krw: string;
  applied_rate: string;
  status: string;
  merchant_name: string | null;
  created_at: string;
}

@Injectable()
export class UsersService {
  constructor(@Inject(DB_POOL) private readonly db: Pool) {}

  async getProfile(userId: string): Promise<UserProfile> {
    const result = await this.db.query<UserProfile>(
      `SELECT id, email, kyc_status, status, created_at
       FROM core.users WHERE id = $1`,
      [userId],
    );
    if (!result.rows[0]) throw new NotFoundException('사용자를 찾을 수 없습니다.');
    return result.rows[0];
  }

  async getWallets(userId: string): Promise<WalletAddress[]> {
    const result = await this.db.query<WalletAddress>(
      `SELECT currency, address, payment_priority,
              CASE currency
                WHEN 'SOL'        THEN 'SOLANA'
                WHEN 'ETH'        THEN 'ETHEREUM'
                WHEN 'USDT_ERC20' THEN 'ETHEREUM'
                WHEN 'USDT_TRC20' THEN 'TRON'
                ELSE 'UNKNOWN'
              END AS network
       FROM core.user_wallets
       WHERE user_id = $1
       ORDER BY payment_priority ASC`,
      [userId],
    );
    return result.rows;
  }

  async getBalances(userId: string): Promise<BalanceEntry[]> {
    // gRPC 대신 직접 DB 쿼리 (MVP: 동일 PostgreSQL 공유)
    // 서비스 분리 이후에는 core gRPC GetUserBalances 호출로 교체
    const result = await this.db.query<{
      currency: string;
      balance: string;
      locked_balance: string;
    }>(
      `SELECT currency,
              balance::text         AS balance,
              locked_balance::text  AS locked_balance
       FROM core.offchain_ledger
       WHERE user_id = $1`,
      [userId],
    );

    return result.rows.map((r) => ({
      currency: r.currency,
      available_balance: r.balance,  // balance - locked_balance는 DB CHECK로 보장
      locked_balance: r.locked_balance,
    }));
  }

  async getTransactions(
    userId: string,
    limit = 50,
    offset = 0,
  ): Promise<TransactionRow[]> {
    const result = await this.db.query<TransactionRow>(
      `SELECT
         t.id,
         t.type,
         t.used_currency      AS used_currency,
         t.used_amount::text  AS used_amount,
         t.amount_krw::text   AS amount_krw,
         t.applied_rate::text AS applied_rate,
         t.status,
         m.business_name      AS merchant_name,
         t.created_at
       FROM core.transactions t
       LEFT JOIN core.merchants m ON m.id = t.merchant_id
       WHERE t.user_id = $1
       ORDER BY t.created_at DESC
       LIMIT $2 OFFSET $3`,
      [userId, limit, offset],
    );
    return result.rows;
  }

  async updatePaymentPriority(
    userId: string,
    priorities: { currency: string; priority: number }[],
  ): Promise<void> {
    const client = await this.db.connect();
    try {
      await client.query('BEGIN');
      for (const { currency, priority } of priorities) {
        await client.query(
          `UPDATE core.user_wallets
           SET payment_priority = $1
           WHERE user_id = $2 AND currency = $3`,
          [priority, userId, currency],
        );
      }
      await client.query('COMMIT');
    } catch (err) {
      await client.query('ROLLBACK');
      throw err;
    } finally {
      client.release();
    }
  }
}
