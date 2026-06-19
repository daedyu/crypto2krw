import {
  Controller,
  Get,
  Put,
  Body,
  UseGuards,
  Query,
  ParseIntPipe,
  DefaultValuePipe,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { UsersService, UserProfile, WalletAddress, BalanceEntry, TransactionRow } from './users.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/strategies/jwt.strategy';
import { IsArray, IsInt, IsString, Min, Max, ValidateNested } from 'class-validator';
import { Type } from 'class-transformer';

class PriorityItem {
  @IsString() currency!: string;
  @IsInt() @Min(1) @Max(10) priority!: number;
}

class UpdatePriorityDto {
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => PriorityItem)
  priorities!: PriorityItem[];
}

@UseGuards(JwtAuthGuard)
@Controller('users')
export class UsersController {
  constructor(private readonly usersService: UsersService) {}

  /**
   * GET /api/v1/users/me
   * 내 프로필 (이메일, KYC 상태, 계정 상태)
   */
  @Get('me')
  async getMe(@CurrentUser() user: JwtPayload): Promise<UserProfile> {
    return this.usersService.getProfile(user.sub);
  }

  /**
   * GET /api/v1/users/me/balances
   * 코인별 오프체인 잔액 (available / locked)
   */
  @Get('me/balances')
  async getBalances(@CurrentUser() user: JwtPayload): Promise<BalanceEntry[]> {
    return this.usersService.getBalances(user.sub);
  }

  /**
   * GET /api/v1/users/me/wallets
   * 코인별 입금 주소 목록 (결제 우선순위 오름차순)
   */
  @Get('me/wallets')
  async getWallets(@CurrentUser() user: JwtPayload): Promise<WalletAddress[]> {
    return this.usersService.getWallets(user.sub);
  }

  /**
   * GET /api/v1/users/me/transactions?limit=50&offset=0
   * 거래 내역 (입금 + 결제, 최신순)
   */
  @Get('me/transactions')
  async getTransactions(
    @CurrentUser() user: JwtPayload,
    @Query('limit', new DefaultValuePipe(50), ParseIntPipe) limit: number,
    @Query('offset', new DefaultValuePipe(0), ParseIntPipe) offset: number,
  ): Promise<TransactionRow[]> {
    return this.usersService.getTransactions(user.sub, Math.min(limit, 100), offset);
  }

  /**
   * PUT /api/v1/users/me/payment-priority
   * 결제 코인 우선순위 변경
   * Body: { priorities: [{ currency: "SOL", priority: 1 }, ...] }
   */
  @Put('me/payment-priority')
  @HttpCode(HttpStatus.OK)
  async updatePaymentPriority(
    @CurrentUser() user: JwtPayload,
    @Body() dto: UpdatePriorityDto,
  ): Promise<{ message: string }> {
    await this.usersService.updatePaymentPriority(user.sub, dto.priorities);
    return { message: '결제 우선순위가 업데이트되었습니다.' };
  }
}
