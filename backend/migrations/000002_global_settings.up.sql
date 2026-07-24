CREATE TABLE IF NOT EXISTS global_settings (
    id INT PRIMARY KEY DEFAULT 1,
    maintenance_mode BOOLEAN NOT NULL DEFAULT FALSE,
    allow_registrations BOOLEAN NOT NULL DEFAULT TRUE,
    gas_fee_percentage NUMERIC NOT NULL DEFAULT 20,
    min_deposit_usdt NUMERIC NOT NULL DEFAULT 500,
    max_active_layers_per_user INT NOT NULL DEFAULT 10,
    support_email VARCHAR(255) NOT NULL DEFAULT 'support@mautrade.com',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row CHECK (id = 1)
);

INSERT INTO global_settings (id) VALUES (1) ON CONFLICT DO NOTHING;
