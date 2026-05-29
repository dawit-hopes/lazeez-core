-- branch_type is owned by merchants; remove redundant column from users
ALTER TABLE users DROP COLUMN IF EXISTS branch_type;
