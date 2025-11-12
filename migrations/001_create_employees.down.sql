DROP INDEX IF EXISTS idx_employees_city;
DROP INDEX IF EXISTS idx_employees_full_name_trgm;
DROP INDEX IF EXISTS idx_employees_phone;
DROP TABLE IF EXISTS employees;
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";
