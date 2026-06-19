import { IsEmail, IsString, MinLength, MaxLength } from 'class-validator';

export class MerchantRegisterDto {
  @IsEmail({}, { message: '올바른 이메일 형식이 아닙니다.' })
  email!: string;

  @IsString()
  @MinLength(8, { message: '비밀번호는 최소 8자 이상이어야 합니다.' })
  @MaxLength(72)
  password!: string;

  @IsString()
  @MinLength(1, { message: '업체명을 입력해주세요.' })
  @MaxLength(100)
  business_name!: string;
}
