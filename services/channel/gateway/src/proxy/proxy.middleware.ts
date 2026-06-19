import { Injectable, NestMiddleware } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { Request, Response, NextFunction } from 'express';
import { createProxyMiddleware, fixRequestBody } from 'http-proxy-middleware';

const AUTH_SERVICE_URL  = process.env.AUTH_SERVICE_URL  ?? 'http://localhost:3001';
const PAYMENT_SERVICE_URL = process.env.PAYMENT_SERVICE_URL ?? 'http://localhost:3003';
const ORACLE_SERVICE_URL  = process.env.ORACLE_SERVICE_URL  ?? 'http://localhost:3002';

// 인증이 필요 없는 경로 패턴
const PUBLIC_PATHS = [
  /^\/api\/v1\/auth\/register/,
  /^\/api\/v1\/auth\/login/,
  /^\/api\/v1\/auth\/refresh/,
  /^\/api\/v1\/auth\/merchant\/register/,
  /^\/api\/v1\/auth\/merchant\/login/,
  /^\/api\/v1\/admin\//,
  /^\/api\/v1\/rates/,
  /^\/health/,
];

function isPublic(path: string): boolean {
  return PUBLIC_PATHS.some((pattern) => pattern.test(path));
}

function resolveTarget(path: string): string | null {
  if (path.startsWith('/api/v1/auth') || path.startsWith('/api/v1/users') || path.startsWith('/api/v1/merchant') || path.startsWith('/api/v1/admin')) {
    return AUTH_SERVICE_URL;
  }
  if (path.startsWith('/api/v1/payment')) {
    return PAYMENT_SERVICE_URL;
  }
  if (path.startsWith('/api/v1/rates')) {
    return ORACLE_SERVICE_URL;
  }
  return null;
}

@Injectable()
export class ProxyMiddleware implements NestMiddleware {
  constructor(private readonly jwtService: JwtService) {}

  use(req: Request, res: Response, next: NextFunction): void {
    const path = req.originalUrl ?? req.url ?? '';

    if (!isPublic(path)) {
      const authHeader = req.headers['authorization'];
      const token = typeof authHeader === 'string' && authHeader.startsWith('Bearer ')
        ? authHeader.slice(7)
        : null;

      if (!token) {
        res.writeHead(401, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ success: false, data: null, error: { code: 'UNAUTHORIZED', message: '인증이 필요합니다.' } }));
        return;
      }

      try {
        this.jwtService.verify(token, { secret: process.env.JWT_ACCESS_SECRET });
      } catch {
        res.writeHead(401, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ success: false, data: null, error: { code: 'UNAUTHORIZED', message: '유효하지 않은 토큰입니다.' } }));
        return;
      }
    }

    const target = resolveTarget(path);
    if (!target) {
      res.writeHead(404, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ success: false, data: null, error: { code: 'NOT_FOUND', message: '경로를 찾을 수 없습니다.' } }));
      return;
    }

    // NestJS가 req.url을 마운트 기준 상대경로로 바꾸므로 원본 경로로 복원
    req.url = path;

    const proxy = createProxyMiddleware({
      target,
      changeOrigin: true,
      on: { proxyReq: fixRequestBody },
    });

    proxy(req, res, next);
  }
}
