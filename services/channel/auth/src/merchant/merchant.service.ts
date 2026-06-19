import {
  Injectable,
  ConflictException,
  UnauthorizedException,
  NotFoundException,
  Logger,
} from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { Inject } from '@nestjs/common';
import { Pool } from 'pg';
import * as bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import { DB_POOL } from '../database/database.module';
import { MerchantRegisterDto } from './dto/merchant-register.dto';

const BCRYPT_ROUNDS = 12;

export interface MerchantTokenPayload {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  merchant_id: string;
  business_name: string;
}

export interface MerchantProfile {
  id: string;
  email: string;
  business_name: string;
  status: string;
  krw_balance: string;
}

interface MerchantRow extends MerchantProfile {
  password_hash: string;
}

@Injectable()
export class MerchantService {
  private readonly logger = new Logger(MerchantService.name);

  constructor(
    @Inject(DB_POOL) private readonly db: Pool,
    private readonly jwtService: JwtService,
  ) {}

  async register(dto: MerchantRegisterDto): Promise<MerchantTokenPayload> {
    const existing = await this.db.query<{ id: string }>(
      `SELECT id FROM core.merchants WHERE email = $1`,
      [dto.email],
    );
    if (existing.rows.length > 0) {
      throw new ConflictException('이미 사용 중인 이메일입니다.');
    }

    const passwordHash = await bcrypt.hash(dto.password, BCRYPT_ROUNDS);
    const merchantId = uuidv4();

    await this.db.query(
      `INSERT INTO core.merchants (id, email, password_hash, business_name, status)
       VALUES ($1, $2, $3, $4, 'ACTIVE')`,
      [merchantId, dto.email, passwordHash, dto.business_name],
    );

    this.logger.log(`merchant registered: ${merchantId}`);
    return this.issueTokens(merchantId, dto.email, dto.business_name);
  }

  async login(email: string, password: string): Promise<MerchantTokenPayload> {
    const result = await this.db.query<MerchantRow>(
      `SELECT id, email, password_hash, business_name, status
       FROM core.merchants WHERE email = $1`,
      [email],
    );
    const merchant = result.rows[0];

    if (!merchant) {
      throw new UnauthorizedException('이메일 또는 비밀번호가 올바르지 않습니다.');
    }

    const valid = await bcrypt.compare(password, merchant.password_hash);
    if (!valid) {
      throw new UnauthorizedException('이메일 또는 비밀번호가 올바르지 않습니다.');
    }

    if (merchant.status === 'SUSPENDED') {
      throw new UnauthorizedException('정지된 계정입니다. 고객센터에 문의해주세요.');
    }

    this.logger.log(`merchant logged in: ${merchant.id}`);
    return this.issueTokens(merchant.id, merchant.email, merchant.business_name);
  }

  async getProfile(merchantId: string): Promise<MerchantProfile> {
    const result = await this.db.query<MerchantProfile>(
      `SELECT id, email, business_name, status, krw_balance
       FROM core.merchants WHERE id = $1`,
      [merchantId],
    );
    if (!result.rows[0]) throw new NotFoundException('가맹점을 찾을 수 없습니다.');
    return result.rows[0];
  }

  async updateBusinessName(merchantId: string, businessName: string): Promise<{ business_name: string }> {
    const result = await this.db.query<{ business_name: string }>(
      `UPDATE core.merchants SET business_name = $1, updated_at = now()
       WHERE id = $2
       RETURNING business_name`,
      [businessName, merchantId],
    );
    if (!result.rows[0]) throw new NotFoundException('가맹점을 찾을 수 없습니다.');
    return result.rows[0];
  }

  private issueTokens(merchantId: string, email: string, businessName: string): MerchantTokenPayload {
    const payload = { sub: merchantId, email, role: 'merchant', businessName };

    const accessToken = this.jwtService.sign(payload, {
      secret: process.env.JWT_ACCESS_SECRET,
      expiresIn: process.env.JWT_ACCESS_EXPIRES_IN ?? '15m',
    });

    const refreshToken = this.jwtService.sign(payload, {
      secret: process.env.JWT_REFRESH_SECRET,
      expiresIn: process.env.JWT_REFRESH_EXPIRES_IN ?? '7d',
    });

    return {
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_in: 15 * 60,
      merchant_id: merchantId,
      business_name: businessName,
    };
  }
}
