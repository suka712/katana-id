ALTER TABLE users ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT NOW();

ALTER TABLE providers ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE providers ALTER COLUMN created_at SET DEFAULT NOW();

ALTER TABLE otps ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE sessions ALTER COLUMN token SET DEFAULT gen_random_uuid();
