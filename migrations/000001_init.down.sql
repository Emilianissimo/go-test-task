BEGIN;

-- 1. Удаляем индексы (хотя DROP TABLE сделает это сам, для чистоты структуры в скрипте полезно)
DROP INDEX IF EXISTS idx_tx_idempotency;
DROP INDEX IF EXISTS idx_payout_wallet_from;
DROP INDEX IF EXISTS idx_wallets_user_id;

-- 2. Удаляем таблицы в порядке, обратном созданию
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS users;

-- 3. Удаляем кастомные типы (ENUM)
DROP TYPE IF EXISTS payout_status;

COMMIT;
