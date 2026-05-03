-- Run once as a superuser (e.g. postgres) on the server that returned:
--   FATAL: role "telemetry" does not exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'telemetry') THEN
    CREATE ROLE telemetry WITH LOGIN PASSWORD 'telemetry';
  END IF;
END$$;
