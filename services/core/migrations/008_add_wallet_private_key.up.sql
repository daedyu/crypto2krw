-- 개발 전용: Solana ed25519 개인키를 DB에 평문 저장.
-- 프로덕션에서는 HSM(Hardware Security Module) 또는 KMS로 교체 필요.
ALTER TABLE core.user_wallets
    ADD COLUMN private_key_hex TEXT;
