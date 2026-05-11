CREATE TABLE IF NOT EXISTS cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

    card_number_encrypted BYTEA NOT NULL,
    expiry_encrypted BYTEA NOT NULL,
    cvv_hash TEXT NOT NULL,

    card_number_hmac TEXT NOT NULL,
    expiry_hmac TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);