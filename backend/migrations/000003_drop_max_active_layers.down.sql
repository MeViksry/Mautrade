ALTER TABLE global_settings ADD COLUMN IF NOT EXISTS max_active_layers_per_user INT NOT NULL DEFAULT 10;
