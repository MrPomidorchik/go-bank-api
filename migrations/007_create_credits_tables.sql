ALTER TABLE transactions
DROP CONSTRAINT IF EXISTS transactions_type_check;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_type_check
        CHECK (type IN (
                        'deposit',
                        'withdraw',
                        'transfer_in',
                        'transfer_out',
                        'card_payment',
                        'credit_disbursement'
            ));

CREATE TABLE IF NOT EXISTS credits (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

    amount NUMERIC(18, 2) NOT NULL CHECK (amount > 0),
    annual_rate NUMERIC(5, 2) NOT NULL CHECK (annual_rate >= 0),
    term_months INT NOT NULL CHECK (term_months > 0),
    monthly_payment NUMERIC(18, 2) NOT NULL CHECK (monthly_payment > 0),
    remaining_amount NUMERIC(18, 2) NOT NULL CHECK (remaining_amount >= 0),

    status VARCHAR(20) NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'closed', 'overdue')),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS payment_schedules (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_id UUID NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

    payment_number INT NOT NULL,
    due_date DATE NOT NULL,

    amount NUMERIC(18, 2) NOT NULL CHECK (amount > 0),
    principal_amount NUMERIC(18, 2) NOT NULL CHECK (principal_amount >= 0),
    interest_amount NUMERIC(18, 2) NOT NULL CHECK (interest_amount >= 0),
    penalty_amount NUMERIC(18, 2) NOT NULL DEFAULT 0 CHECK (penalty_amount >= 0),

    status VARCHAR(20) NOT NULL DEFAULT 'planned'
    CHECK (status IN ('planned', 'paid', 'overdue')),

    paid_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);
CREATE INDEX IF NOT EXISTS idx_credits_account_id ON credits(account_id);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_credit_id ON payment_schedules(credit_id);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_user_id ON payment_schedules(user_id);