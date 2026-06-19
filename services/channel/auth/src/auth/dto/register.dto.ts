import { IsEmail, IsString, MinLength, MaxLength } from 'class-validator';

export class RegisterDto {
  @IsEmail({}, { message: '올바른 이메일 형식이 아닙니다.' })
  email!: string;

  @IsString()
  @MinLength(8, { message: '비밀번호는 최소 8자 이상이어야 합니다.' })
  @MaxLength(72, { message: '비밀번호는 최대 72자까지 가능합니다.' })
  password!: string;
}
