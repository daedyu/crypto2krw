.PHONY: infra infra-down dev-core dev-oracle dev-payment dev-auth dev-gateway dev-merchant stop

COMPOSE = docker compose -f infra/docker-compose.yml

# 인프라(Postgres, Redis, Kafka) 시작
infra:
	$(COMPOSE) up -d
	@echo "인프라 기동 완료. 마이그레이션: make migrate"

infra-down:
	$(COMPOSE) down

# 데이터 포함 완전 초기화 (Kafka NodeExistsException 발생 시 사용)
infra-clean:
	$(COMPOSE) down -v
	$(COMPOSE) up -d

# DB 마이그레이션
migrate:
	migrate -database "postgres://crypto2krw:crypto2krw_dev@localhost:5432/crypto2krw?sslmode=disable" \
	        -path services/core/migrations up

# 빌드
build-core:
	go -C services/core build -o /tmp/core-server ./cmd/server

build-oracle:
	go -C services/oracle build -o /tmp/oracle-server ./cmd/oracle

build-payment:
	go -C services/payment build -o /tmp/payment-server ./cmd/payment

build-auth:
	cd services/channel/auth && npm run build

build-gateway:
	cd services/channel/gateway && npm run build

build: build-core build-oracle build-payment build-auth build-gateway
	@echo "전체 빌드 완료"

# 개발 서버 실행 (각 서비스별 개별 실행)
dev-core:
	go -C services/core run ./cmd/server

dev-oracle:
	go -C services/oracle run ./cmd/oracle

dev-payment:
	cd services/payment && go run ./cmd/payment

dev-auth:
	cd services/channel/auth && npm run start:dev

dev-gateway:
	cd services/channel/gateway && npm run start:dev

dev-merchant:
	cd apps/MerchantWeb && npm run dev

# 타입체크
typecheck:
	cd services/channel/auth && npx tsc --noEmit
	cd services/channel/gateway && npx tsc --noEmit
	go -C services/core vet ./...
	go -C services/oracle vet ./...
	go -C services/payment vet ./...
