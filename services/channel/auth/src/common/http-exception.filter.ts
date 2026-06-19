import {
  ExceptionFilter,
  Catch,
  ArgumentsHost,
  HttpException,
  HttpStatus,
  Logger,
} from '@nestjs/common';
import { FastifyReply } from 'fastify';

@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  private readonly logger = new Logger(HttpExceptionFilter.name);

  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const reply = ctx.getResponse<FastifyReply>();

    let status = HttpStatus.INTERNAL_SERVER_ERROR;
    let code = 'INTERNAL_ERROR';
    let message = '서버 내부 오류가 발생했습니다.';

    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const res = exception.getResponse();

      if (typeof res === 'object' && res !== null) {
        const body = res as Record<string, unknown>;
        code = (body['error'] as string) ?? HttpStatus[status];
        // class-validator 배열 메시지 처리
        const msg = body['message'];
        message = Array.isArray(msg) ? msg.join(', ') : (msg as string) ?? message;
      } else {
        message = res as string;
        code = HttpStatus[status];
      }
    } else {
      this.logger.error('Unhandled exception', exception);
    }

    void reply.status(status).send({
      success: false,
      data: null,
      error: { code, message },
    });
  }
}
