import { Controller, Post, Get, Body, Query, HttpCode, HttpStatus } from '@nestjs/common';
import { IsString, IsNumberString, IsIn } from 'class-validator';
import { AdminService, DepositResult, UserBalance } from './admin.service';

class DepositDto {
  @IsString() email!: string;
  @IsIn(['SOL', 'ETH', 'USDT_TRC20']) currency!: string;
  @IsNumberString() amount!: string;
}

@Controller('admin')
export class AdminController {
  constructor(private readonly adminService: AdminService) {}

  /** POST /api/v1/admin/deposit */
  @Post('deposit')
  @HttpCode(HttpStatus.OK)
  deposit(@Body() dto: DepositDto): Promise<DepositResult> {
    return this.adminService.deposit(dto.email, dto.currency, dto.amount);
  }

  /** GET /api/v1/admin/balances?email=... */
  @Get('balances')
  getBalances(@Query('email') email: string): Promise<UserBalance> {
    return this.adminService.getBalances(email);
  }
}
