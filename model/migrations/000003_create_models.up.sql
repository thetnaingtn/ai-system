BEGIN;

CREATE SCHEMA model_catalog;

CREATE TYPE model_catalog.approval_status AS ENUM (
  'pending',
  'approved',
  'rejected'
);

CREATE TYPE model_catalog.accounting_mode AS ENUM (
  'metered',
  'unmetered'
);

CREATE TABLE model_catalog.models (
  name text PRIMARY KEY,
  provider text NOT NULL,
  base_url text NOT NULL,
  credential_secret_ref text,
  owner_tenant_id uuid REFERENCES tenant_config.tenants(id) ON DELETE RESTRICT,
  capabilities jsonb NOT NULL,
  status text NOT NULL DEFAULT 'active',
  sunset_at timestamptz,
  replacement_model text REFERENCES model_catalog.models(name) ON DELETE RESTRICT,
  accounting_mode model_catalog.accounting_mode NOT NULL,
  approval_status model_catalog.approval_status NOT NULL DEFAULT 'pending',
  config_version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT models_name_valid
    CHECK (name <> '' AND name = btrim(name)),
  CONSTRAINT models_provider_valid
    CHECK (provider <> '' AND provider = btrim(provider)),
  CONSTRAINT models_base_url_valid
    CHECK (base_url <> '' AND base_url = btrim(base_url)),
  CONSTRAINT models_secret_ref_valid
    CHECK (credential_secret_ref IS NULL OR
      (credential_secret_ref <> '' AND
       credential_secret_ref = btrim(credential_secret_ref))),
  CONSTRAINT models_capabilities_object
    CHECK (jsonb_typeof(capabilities) = 'object'),
  CONSTRAINT models_status_valid
    CHECK (status IN ('active', 'deprecated', 'sunset')),
  CONSTRAINT models_replacement_not_self
    CHECK (replacement_model IS NULL OR replacement_model <> name),
  CONSTRAINT models_config_version_positive
    CHECK (config_version > 0)
);

CREATE INDEX models_owner_tenant_id_idx
  ON model_catalog.models(owner_tenant_id);

CREATE INDEX models_replacement_model_idx
  ON model_catalog.models(replacement_model);

COMMIT;
