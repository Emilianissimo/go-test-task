BEGIN;

DROP INDEX IF EXISTS idx_transactions_op_composite;
DROP INDEX IF EXISTS idx_transactions_tx_id;
DROP INDEX IF EXISTS idx_payouts_tx_id;
DROP INDEX IF EXISTS idx_payout_wallet_from;
DROP INDEX IF EXISTS idx_wallets_user_id;

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS payout_status;

COMMIT;
