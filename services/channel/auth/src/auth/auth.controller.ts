import {
  Controller,
  Post,
  Body,
  HttpCode,
  HttpStatus,
  UseGuards,
} from '@nestjs/common';
import { AuthService, TokenPair } from './auth.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { RefreshDto } from './dto/refresh.dto';
import { JwtAuthGuard } from './guards/jwt-auth.guard';
import { CurrentUser } from './decorators/current-user.decorator';
import { JwtPayload } from './strategies/jwt.strategy';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  /**
   * POST /api/v1/auth/register
   * 신규 회원가입. 성공 시 access/refresh 토큰 즉시 반환.
   * Kafka user.registered 이벤트를 발행해 core 서비스가 지갑을 비동기 생성.
   */
  @Post('register')
  @HttpCode(HttpStatus.CREATED)
  async register(@Body() dto: RegisterDto): Promise<TokenPair> {
    return this.authService.register(dto);
  }

  /**
   * POST /api/v1/auth/login
   * 이메일/비밀번호 로그인.
   */
  @Post('login')
  @HttpCode(HttpStatus.OK)
  async login(@Body() dto: LoginDto): Promise<TokenPair> {
    return this.authService.login(dto);
  }

  /**
   * POST /api/v1/auth/refresh
   * 리프레시 토큰으로 새 access 토큰 발급 (토큰 로테이션).
   */
  @Post('refresh')
  @HttpCode(HttpStatus.OK)
  async refresh(@Body() dto: RefreshDto): Promise<TokenPair> {
    return this.authService.refresh(dto.refresh_token);
  }

  /**
   * POST /api/v1/auth/logout
   * JWT 인증 필요. 리프레시 토큰 DB 삭제.
   */
  @UseGuards(JwtAuthGuard)
  @Post('logout')
  @HttpCode(HttpStatus.OK)
  async logout(@CurrentUser() user: JwtPayload): Promise<{ message: string }> {
    await this.authService.logout(user.sub);
    return { message: '로그아웃되었습니다.' };
  }
}
