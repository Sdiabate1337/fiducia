-- Add onboarding_completed to cabinets
ALTER TABLE cabinets ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE;
