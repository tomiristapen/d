ALTER TABLE profiles
ADD COLUMN IF NOT EXISTS activity_level TEXT NOT NULL DEFAULT 'moderate';

