import {
  Controller,
  Post,
  Get,
  Patch,
  Body,
  HttpCode,
  HttpStatus,
  UseGuards,
} from '@nestjs/common';
import { IsString, MinLength, MaxLength } from 'class-validator';
import { MerchantService, MerchantTokenPayload, MerchantProfile } from './merchant.service';
import { MerchantRegisterDto } from './dto/merchant-register.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/strategies/jwt.strategy';

class MerchantLoginDto {
  @IsString() email!: string;
  @IsString() password!: string;
}

class UpdateBusinessNameDto {
  @IsString()
  @MinLength(1, { message: '업체명을 입력해주세요.' })
  @MaxLength(100)
  business_name!: string;
}

@Controller()
export class MerchantController {
  constructor(private readonly merchantService: MerchantService) {}

  /** POST /api/v1/auth/merchant/register */
  @Post('auth/merchant/register')
  @HttpCode(HttpStatus.CREATED)
  register(@Body() dto: MerchantRegisterDto): Promise<MerchantTokenPayload> {
    return this.merchantService.register(dto);
  }

  /** POST /api/v1/auth/merchant/login */
  @Post('auth/merchant/login')
  @HttpCode(HttpStatus.OK)
  login(@Body() dto: MerchantLoginDto): Promise<MerchantTokenPayload> {
    return this.merchantService.login(dto.email, dto.password);
  }

  /** GET /api/v1/merchant/me */
  @Get('merchant/me')
  @UseGuards(JwtAuthGuard)
  getMe(@CurrentUser() user: JwtPayload): Promise<MerchantProfile> {
    return this.merchantService.getProfile(user.sub);
  }

  /** PATCH /api/v1/merchant/me */
  @Patch('merchant/me')
  @UseGuards(JwtAuthGuard)
  @HttpCode(HttpStatus.OK)
  updateBusinessName(
    @CurrentUser() user: JwtPayload,
    @Body() dto: UpdateBusinessNameDto,
  ) {
    return this.merchantService.updateBusinessName(user.sub, dto.business_name);
  }
}
