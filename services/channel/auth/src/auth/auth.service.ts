import {
  Injectable,
  ConflictException,
  UnauthorizedException,
  Logger,
} from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { Inject } from '@nestjs/common';
import { Pool } from 'pg';
import * as bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import { DB_POOL } from '../database/database.module';
import { KafkaService } from '../kafka/kafka.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { JwtPayload } from './strategies/jwt.strategy';

const BCRYPT_ROUNDS = 12;
const REFRESH_TOKEN_EXPIRES_DAYS = 7;

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;  // seconds
}

interface UserRow {
  id: string;
  email: string;
  password_hash: string;
  kyc_status: string;
  status: string;
  created_at: Date;
}

@Injectable()
export class AuthService {
  private readonly logger = new Logger(AuthService.name);

  constructor(
    @Inject(DB_POOL) private readonly db: Pool,
    private readonly jwtService: JwtService,
    private readonly kafkaService: KafkaService,
  ) {}

  async register(dto: RegisterDto): Promise<TokenPair> {
    const existing = await this.db.query<{ id: string }>(
      `SELECT id FROM core.users WHERE email = $1`,
      [dto.email],
    );
    if (existing.rows.length > 0) {
      throw new ConflictException('이미 사용 중인 이메일입니다.');
    }

    const passwordHash = await bcrypt.hash(dto.password, BCRYPT_ROUNDS);

    const result = await this.db.query<UserRow>(
      `INSERT INTO core.users (email, password_hash)
       VALUES ($1, $2)
       RETURNING id, email, kyc_status, status, created_at`,
      [dto.email, passwordHash],
    );
    const user = result.rows[0];

    // 비동기 지갑 생성 트리거 — core 서비스가 이벤트를 수신해 처리
    await this.kafkaService.publishUserRegistered(user.id, user.email);

    this.logger.log(`user registered: ${user.id}`);
    return this.issueTokenPair(user.id, user.email);
  }

  async login(dto: LoginDto): Promise<TokenPair> {
    const result = await this.db.query<UserRow>(
      `SELECT id, email, password_hash, status FROM core.users WHERE email = $1`,
      [dto.email],
    );
    const user = result.rows[0];

    if (!user) throw new UnauthorizedException('이메일 또는 비밀번호가 올바르지 않습니다.');

    const valid = await bcrypt.compare(dto.password, user.password_hash);
    if (!valid) throw new UnauthorizedException('이메일 또는 비밀번호가 올바르지 않습니다.');

    if (user.status === 'SUSPENDED') {
      throw new UnauthorizedException('정지된 계정입니다. 고객센터에 문의해주세요.');
    }

    this.logger.log(`user logged in: ${user.id}`);
    return this.issueTokenPair(user.id, user.email);
  }

  async refresh(rawRefreshToken: string): Promise<TokenPair> {
    let payload: JwtPayload;
    try {
      payload = this.jwtService.verify<JwtPayload>(rawRefreshToken, {
        secret: process.env.JWT_REFRESH_SECRET,
      });
    } catch {
      throw new UnauthorizedException('유효하지 않거나 만료된 리프레시 토큰입니다.');
    }

    const tokenHash = await bcrypt.hash(rawRefreshToken, 6);

    // 저장된 리프레시 토큰 검증
    const stored = await this.db.query<{ id: string; expires_at: Date }>(
      `SELECT id, expires_at FROM core.refresh_tokens
       WHERE user_id = $1 AND expires_at > now()
       ORDER BY created_at DESC LIMIT 1`,
      [payload.sub],
    );

    if (stored.rows.length === 0) {
      throw new UnauthorizedException('리프레시 토큰이 만료되었거나 존재하지 않습니다.');
    }

    // 기존 토큰 삭제 후 새 토큰 발급 (rotation)
    await this.db.query(`DELETE FROM core.refresh_tokens WHERE user_id = $1`, [payload.sub]);

    this.logger.log(`token refreshed for user: ${payload.sub}`);
    return this.issueTokenPair(payload.sub, payload.email);
  }

  async logout(userId: string): Promise<void> {
    await this.db.query(`DELETE FROM core.refresh_tokens WHERE user_id = $1`, [userId]);
    this.logger.log(`user logged out: ${userId}`);
  }

  private async issueTokenPair(userId: string, email: string): Promise<TokenPair> {
    const payload: JwtPayload = { sub: userId, email };

    const accessToken = this.jwtService.sign(payload, {
      secret: process.env.JWT_ACCESS_SECRET,
      expiresIn: process.env.JWT_ACCESS_EXPIRES_IN ?? '15m',
    });

    const refreshToken = this.jwtService.sign(payload, {
      secret: process.env.JWT_REFRESH_SECRET,
      expiresIn: process.env.JWT_REFRESH_EXPIRES_IN ?? '7d',
    });

    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + REFRESH_TOKEN_EXPIRES_DAYS);

    // refresh_tokens 테이블에 저장 (해시로 보관)
    const tokenId = uuidv4();
    const tokenHash = await bcrypt.hash(refreshToken, 6);  // 낮은 cost — 빠른 저장용
    await this.db.query(
      `INSERT INTO core.refresh_tokens (id, user_id, token_hash, expires_at)
       VALUES ($1, $2, $3, $4)`,
      [tokenId, userId, tokenHash, expiresAt],
    );

    return {
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_in: 15 * 60,  // 15분 (초 단위)
    };
  }
}
