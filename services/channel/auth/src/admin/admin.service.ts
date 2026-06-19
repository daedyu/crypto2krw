import { Injectable, NotFoundException, BadRequestException, Inject } from '@nestjs/common';
import { Pool } from 'pg';
import { DB_POOL } from '../database/database.module';

const VALID_CURRENCIES = ['SOL', 'ETH', 'USDT_TRC20'] as const;
type Currency = typeof VALID_CURRENCIES[number];

export interface DepositResult {
  user_id: string;
  email: string;
  currency: string;
  deposited: string;
  new_balance: string;
}

export interface UserBalance {
  user_id: string;
  email: string;
  balances: { currency: string; balance: string }[];
}

@Injectable()
export class AdminService {
  constructor(@Inject(DB_POOL) private readonly db: Pool) {}

  async deposit(email: string, currency: string, amount: string): Promise<DepositResult> {
    if (!(VALID_CURRENCIES as readonly string[]).includes(currency)) {
      throw new BadRequestException(`지원 통화: ${VALID_CURRENCIES.join(', ')}`);
    }

    const amt = parseFloat(amount);
    if (isNaN(amt) || amt <= 0) {
      throw new BadRequestException('금액은 0보다 커야 합니다.');
    }

    const userRow = await this.db.query<{ id: string; email: string }>(
      `SELECT id, email FROM core.users WHERE email = $1`,
      [email],
    );
    if (!userRow.rows[0]) {
      throw new NotFoundException(`사용자를 찾을 수 없습니다: ${email}`);
    }
    const { id: userId } = userRow.rows[0];

    const result = await this.db.query<{ balance: string }>(
      `INSERT INTO core.offchain_ledger (user_id, currency, balance)
       VALUES ($1, $2, $3)
       ON CONFLICT (user_id, currency)
       DO UPDATE SET
         balance    = core.offchain_ledger.balance + EXCLUDED.balance,
         updated_at = now()
       RETURNING balance`,
      [userId, currency as Currency, amount],
    );

    return {
      user_id: userId,
      email,
      currency,
      deposited: amount,
      new_balance: result.rows[0].balance,
    };
  }

  async getBalances(email: string): Promise<UserBalance> {
    const userRow = await this.db.query<{ id: string; email: string }>(
      `SELECT id, email FROM core.users WHERE email = $1`,
      [email],
    );
    if (!userRow.rows[0]) {
      throw new NotFoundException(`사용자를 찾을 수 없습니다: ${email}`);
    }
    const { id: userId } = userRow.rows[0];

    const ledger = await this.db.query<{ currency: string; balance: string }>(
      `SELECT currency, balance FROM core.offchain_ledger WHERE user_id = $1 ORDER BY currency`,
      [userId],
    );

    return { user_id: userId, email, balances: ledger.rows };
  }
}
