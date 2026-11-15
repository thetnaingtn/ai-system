BEGIN;

CREATE SCHEMA tenant_config;

CREATE TYPE tenant_config.tenant_tier AS ENUM (
  'experimental',
  'production'
);

CREATE TABLE tenant_config.tenants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  identity_key text NOT NULL UNIQUE,
  owner text,
  tier tenant_config.tenant_tier NOT NULL DEFAULT 'experimental',
  review_by date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT tenants_identity_key_valid
    CHECK (identity_key <> '' AND identity_key = btrim(identity_key)),
  CONSTRAINT tenants_owner_valid
    CHECK (owner IS NULL OR (owner <> '' AND owner = btrim(owner))),
  CONSTRAINT tenants_production_owner_required
    CHECK (tier <> 'production' OR owner IS NOT NULL)
);

COMMIT;
