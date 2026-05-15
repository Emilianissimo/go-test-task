BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payouts (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    target_id VARCHAR(255) NOT NULL,
    wallet_from INTEGER NOT NULL REFERENCES wallets(id),
    amount DECIMAL(20, 2) NOT NULL,
    tx_id VARCHAR(20) UNIQUE NOT NULL,
    balance_before DECIMAL(20, 2),
    balance_after DECIMAL(20, 2),
    status SMALLINT NOT NULL DEFAULT 1,    -- 0: created, 1: pending, 2: confirmed, -1: rejected
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    op_type SMALLINT NOT NULL DEFAULT 1, -- 1: payout, 2: deposit (future use)
    op_id INTEGER NOT NULL,               -- ID  payouts/deposits
    tx_id VARCHAR(20) UNIQUE NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,    -- 0: created, 1: pending, 2: confirmed, -1: rejected
    -- Future:
    -- currency VARCHAR(10) DEFAULT 'USD',
    -- provider VARCHAR(50) DEFAULT 'internal',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_payout_wallet_from ON payouts(wallet_from);
CREATE INDEX idx_payouts_tx_id ON payouts(tx_id);
CREATE INDEX idx_transactions_tx_id ON transactions(tx_id);
CREATE INDEX idx_transactions_op_composite ON transactions(op_id, op_type);

-- deposits IS 'Future table for deposits. Will link to transactions with op_type=2';




--- ONLY FOR TESTING PURPOSES! ---
--- SEEDING ---

INSERT INTO users (id, name)
VALUES (1, 'Builder Engineer')
    ON CONFLICT DO NOTHING;

INSERT INTO wallets (uuid, user_id, balance, created_at, updated_at)
VALUES (
       '550e8400-e29b-41d4-a716-446655440000',
       1,
       1000.00,
       NOW(),
       NOW()
   )
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO transactions (
    op_type,
    op_id,
    amount,
    status,
    tx_id
)
VALUES (
       2,                       -- Deposit
       0,                       -- Placeholder
       1000.00,
       2,                    -- Status: confirmed
        'fake'
       );

COMMIT;
