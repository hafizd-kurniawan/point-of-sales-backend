-- Create installments table for credit payment tracking
-- First create enum type for installment status
CREATE TYPE installment_status AS ENUM ('pending', 'paid', 'overdue', 'waived');

CREATE TABLE IF NOT EXISTS installments (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES sales_transactions(id) ON DELETE CASCADE,
    installment_number INTEGER NOT NULL,
    due_date DATE NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    paid_amount DECIMAL(15,2) DEFAULT 0 CHECK (paid_amount >= 0),
    status installment_status NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMP WITH TIME ZONE,
    payment_method VARCHAR(20) CHECK (payment_method IN ('cash', 'transfer', 'check', 'mixed')),
    payment_reference VARCHAR(100),
    notes TEXT,
    waived_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_installment_per_transaction UNIQUE (transaction_id, installment_number),
    CONSTRAINT valid_paid_amount CHECK (paid_amount <= amount)
);

-- Indexes for performance
CREATE INDEX idx_installments_transaction ON installments(transaction_id);
CREATE INDEX idx_installments_status ON installments(status);
CREATE INDEX idx_installments_due_date ON installments(due_date);
CREATE INDEX idx_installments_overdue ON installments(due_date, status) WHERE status = 'pending';

-- Function to automatically update installment status to overdue
CREATE OR REPLACE FUNCTION update_overdue_installments()
RETURNS void AS $$
BEGIN
    UPDATE installments 
    SET status = 'overdue', updated_at = CURRENT_TIMESTAMP
    WHERE status = 'pending' 
    AND due_date < CURRENT_DATE;
END;
$$ LANGUAGE plpgsql;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_installment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_installment_updated_at
    BEFORE UPDATE ON installments
    FOR EACH ROW
    EXECUTE FUNCTION update_installment_updated_at();

-- Function to automatically create installments when a credit sales transaction is created
CREATE OR REPLACE FUNCTION create_installments_for_credit_transaction()
RETURNS TRIGGER AS $$
DECLARE
    remaining_amount DECIMAL(15,2);
    monthly_amount DECIMAL(15,2);
    installment_count INTEGER := 12; -- Default 12 months
    i INTEGER;
BEGIN
    -- Only create installments for credit payment method
    IF NEW.payment_method = 'credit' AND NEW.status = 'completed' THEN
        -- Calculate remaining amount after down payment (assume 20% down payment)
        remaining_amount := NEW.total_amount * 0.8;
        monthly_amount := remaining_amount / installment_count;
        
        -- Create installments
        FOR i IN 1..installment_count LOOP
            INSERT INTO installments (
                transaction_id,
                installment_number,
                due_date,
                amount,
                status
            ) VALUES (
                NEW.id,
                i,
                NEW.transaction_date::DATE + (i || ' months')::INTERVAL,
                monthly_amount,
                'pending'
            );
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_installments_for_credit
    AFTER INSERT ON sales_transactions
    FOR EACH ROW
    EXECUTE FUNCTION create_installments_for_credit_transaction();