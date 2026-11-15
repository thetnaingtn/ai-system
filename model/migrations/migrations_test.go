package migrations_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMigrations uses a disposable PostgreSQL container without published ports
// or host volumes. It never connects to the development database.
func TestMigrations(t *testing.T) {
	if os.Getenv("MODEL_MIGRATION_TEST") != "1" {
		t.Skip("run make test-migrations to verify migrations against PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	name := fmt.Sprintf("ai-system-migration-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	run := func(args ...string) string {
		t.Helper()
		output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("docker %s failed: %v\n%s", strings.Join(args, " "), err, output)
		}
		return strings.TrimSpace(string(output))
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if output, err := exec.CommandContext(cleanupCtx, "docker", "rm", "--force", name).CombinedOutput(); err != nil {
			t.Errorf("remove test PostgreSQL container: %v\n%s", err, output)
		}
	})
	run("run", "--detach", "--rm", "--name", name,
		"--env", "POSTGRES_USER=model", "--env", "POSTGRES_DB=migration_test",
		"--env", "POSTGRES_HOST_AUTH_METHOD=trust", "postgres:18-alpine")
	readyCtx, readyCancel := context.WithTimeout(ctx, 30*time.Second)
	defer readyCancel()
	for {
		if exec.CommandContext(readyCtx, "docker", "exec", name,
			"pg_isready", "-h", "127.0.0.1", "-U", "model", "-d", "migration_test").Run() == nil {
			break
		}
		select {
		case <-readyCtx.Done():
			t.Fatal("test PostgreSQL did not become ready")
		case <-time.After(200 * time.Millisecond):
		}
	}
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	migrate := func(args ...string) {
		t.Helper()
		command := []string{"run", "--rm", "--network", "container:" + name,
			"--volume", dir + ":/migrations:ro", "migrate/migrate:v4.20.1",
			"-path=/migrations", "-database=postgres://model@127.0.0.1:5432/migration_test?sslmode=disable&search_path=public"}
		run(append(command, args...)...)
	}
	psqlArgs := []string{"exec", name, "psql", "-X", "-U", "model", "-d", "migration_test",
		"--set=ON_ERROR_STOP=1", "--set=VERBOSITY=verbose", "-At", "-c"}
	query := func(sql string) string {
		t.Helper()
		return run(append(psqlArgs, sql)...)
	}
	assertQuery := func(sql, want string) {
		t.Helper()
		if got := query(sql); got != want {
			t.Fatalf("query %s: got %q, want %q", sql, got, want)
		}
	}
	reject := func(t *testing.T, sql, sqlState, constraint string) {
		t.Helper()
		output, err := exec.CommandContext(ctx, "docker", append(psqlArgs, sql)...).CombinedOutput()
		if err == nil {
			t.Fatalf("invalid SQL succeeded: %s", sql)
		}
		if !strings.Contains(string(output), sqlState) || !strings.Contains(string(output), constraint) {
			t.Fatalf("expected SQLSTATE %s and constraint %q, got:\n%s", sqlState, constraint, output)
		}
	}

	// Exercise upgrades from the existing bootstrap migration as well as fresh installs.
	migrate("up", "1")
	assertQuery("SELECT version::text || ':' || dirty::text FROM public.schema_migrations", "1:false")
	migrate("up")
	assertQuery("SELECT version::text || ':' || dirty::text FROM public.schema_migrations", "3:false")
	assertQuery("SELECT count(*) FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE (n.nspname, t.typname) IN (('tenant_config', 'tenant_tier'), ('model_catalog', 'approval_status'), ('model_catalog', 'accounting_mode'))", "3")
	query("INSERT INTO tenant_config.tenants (identity_key) VALUES ('team-a')")
	assertQuery("SELECT (id IS NOT NULL AND tier = 'experimental' AND owner IS NULL AND review_by IS NULL AND created_at = updated_at)::text FROM tenant_config.tenants WHERE identity_key = 'team-a'", "true")
	query("INSERT INTO tenant_config.tenants (identity_key, owner, tier) VALUES ('team-b', 'team-b@example.test', 'production')")
	capabilities := `'{"tools":true,"structured_output":false,"streaming":true,"vision":false,"max_context_tokens":128000}'::jsonb`
	insertModel := func(model, extraColumns, extraValues string) string {
		return "INSERT INTO model_catalog.models (name, provider, base_url, capabilities, accounting_mode" + extraColumns + ") VALUES ('" + model + "', 'openai', 'https://example.test', " + capabilities + ", 'metered'" + extraValues + ")"
	}
	query(insertModel("gpt-4o", ", owner_tenant_id", ", (SELECT id FROM tenant_config.tenants WHERE identity_key = 'team-a')"))
	assertQuery("SELECT (status = 'active' AND approval_status = 'pending' AND config_version = 1 AND created_at = updated_at)::text FROM model_catalog.models WHERE name = 'gpt-4o'", "true")
	query(insertModel("claude-sonnet", ", replacement_model", ", 'gpt-4o'"))
	query("INSERT INTO model_catalog.models (name, provider, base_url, capabilities, accounting_mode, approval_status) VALUES ('llama-70b', 'openai', 'https://example.test', " + capabilities + ", 'unmetered', 'approved')")
	assertQuery("SELECT (owner_tenant_id IS NULL AND credential_secret_ref IS NULL)::text FROM model_catalog.models WHERE name = 'llama-70b'", "true")

	tests := []struct {
		name, sql, state, constraint string
	}{
		{"duplicate tenant", "INSERT INTO tenant_config.tenants (identity_key) VALUES ('team-a')", "23505", "tenants_identity_key_key"},
		{"empty identity", "INSERT INTO tenant_config.tenants (identity_key) VALUES ('')", "23514", "tenants_identity_key_valid"},
		{"padded identity", "INSERT INTO tenant_config.tenants (identity_key) VALUES (' team-c')", "23514", "tenants_identity_key_valid"},
		{"invalid tier", "UPDATE tenant_config.tenants SET tier = 'unknown' WHERE identity_key = 'team-a'", "22P02", "tenant_tier"},
		{"production owner required", "UPDATE tenant_config.tenants SET tier = 'production' WHERE identity_key = 'team-a'", "23514", "tenants_production_owner_required"},
		{"blank owner", "UPDATE tenant_config.tenants SET owner = '' WHERE identity_key = 'team-b'", "23514", "tenants_owner_valid"},
		{"duplicate model", insertModel("gpt-4o", "", ""), "23505", "models_pkey"},
		{"empty name", insertModel("", "", ""), "23514", "models_name_valid"},
		{"padded name", insertModel(" gpt-4o", "", ""), "23514", "models_name_valid"},
		{"unknown owner", insertModel("unknown-owner", ", owner_tenant_id", ", '00000000-0000-0000-0000-000000000000'"), "23503", "models_owner_tenant_id_fkey"},
		{"owner delete restricted", "DELETE FROM tenant_config.tenants WHERE identity_key = 'team-a'", "23001", "models_owner_tenant_id_fkey"},
		{"unknown replacement", insertModel("unknown-replacement", ", replacement_model", ", 'missing-model'"), "23503", "models_replacement_model_fkey"},
		{"self replacement", "UPDATE model_catalog.models SET replacement_model = name WHERE name = 'gpt-4o'", "23514", "models_replacement_not_self"},
		{"replacement delete restricted", "DELETE FROM model_catalog.models WHERE name = 'gpt-4o'", "23001", "models_replacement_model_fkey"},
		{"invalid lifecycle", "UPDATE model_catalog.models SET status = 'unknown' WHERE name = 'gpt-4o'", "23514", "models_status_valid"},
		{"invalid approval", "UPDATE model_catalog.models SET approval_status = 'unknown' WHERE name = 'gpt-4o'", "22P02", "approval_status"},
		{"invalid accounting mode", "UPDATE model_catalog.models SET accounting_mode = 'unknown' WHERE name = 'gpt-4o'", "22P02", "accounting_mode"},
		{"missing accounting mode", "INSERT INTO model_catalog.models (name, provider, base_url, capabilities) VALUES ('missing-mode', 'openai', 'https://example.test', " + capabilities + ")", "23502", "accounting_mode"},
		{"invalid version", "UPDATE model_catalog.models SET config_version = 0 WHERE name = 'gpt-4o'", "23514", "models_config_version_positive"},
		{"nonobject capabilities", "UPDATE model_catalog.models SET capabilities = '[]'::jsonb WHERE name = 'gpt-4o'", "23514", "models_capabilities_object"},
		{"blank provider", "UPDATE model_catalog.models SET provider = '' WHERE name = 'gpt-4o'", "23514", "models_provider_valid"},
		{"blank endpoint", "UPDATE model_catalog.models SET base_url = '' WHERE name = 'gpt-4o'", "23514", "models_base_url_valid"},
		{"blank secret reference", "UPDATE model_catalog.models SET credential_secret_ref = '' WHERE name = 'gpt-4o'", "23514", "models_secret_ref_valid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reject(t, test.sql, test.state, test.constraint)
		})
	}
	if t.Failed() {
		t.FailNow()
	}

	migrate("up")
	assertQuery("SELECT count(*) FROM model_catalog.models", "3")
	migrate("down", "1")
	assertQuery("SELECT to_regclass('model_catalog.models') IS NULL", "t")
	assertQuery("SELECT count(*) FROM tenant_config.tenants", "2")
	assertQuery("SELECT count(*) FROM pg_namespace WHERE nspname = 'model_catalog'", "0")
	migrate("up")
	query(insertModel("gpt-4o", ", owner_tenant_id", ", (SELECT id FROM tenant_config.tenants WHERE identity_key = 'team-a')"))
	migrate("down", "2")
	assertQuery("SELECT version::text || ':' || dirty::text FROM public.schema_migrations", "1:false")
	assertQuery("SELECT count(*) FROM pg_namespace WHERE nspname IN ('tenant_config', 'model_catalog')", "0")
	migrate("down", "1")
	assertQuery("SELECT count(*) FROM public.schema_migrations", "0")
	assertQuery("SELECT count(*) FROM pg_namespace WHERE nspname = 'model'", "0")
	migrate("up")
	assertQuery("SELECT version::text || ':' || dirty::text FROM public.schema_migrations", "3:false")
	assertQuery("SELECT count(*) FROM tenant_config.tenants", "0")
	assertQuery("SELECT count(*) FROM model_catalog.models", "0")
}
