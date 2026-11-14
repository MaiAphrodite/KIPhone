# Supabase Connection Notes

Project ref: `bzguwmwapgodhayhwsfr`

## 1. MCP tools available (server: `supabase`)

- `supabase_apply_migration`
- `supabase_create_branch`
- `supabase_create_notebook`
- `supabase_delete_branch`
- `supabase_delete_feedback`
- `supabase_delete_notebook`
- `supabase_deploy_edge_function`
- `supabase_execute_sql`
- `supabase_generate_typescript_types`
- `supabase_get_advisors`
- `supabase_get_edge_function`
- `supabase_get_notebook`
- `supabase_get_project_url`
- `supabase_get_publishable_keys`
- `supabase_get_storage_config`
- `supabase_list_branches`
- `supabase_list_edge_functions`
- `supabase_list_extensions`
- `supabase_list_migrations`
- `supabase_list_notebooks`
- `supabase_list_storage_buckets`
- `supabase_list_tables`
- `supabase_merge_branch`
- `supabase_query_logs`
- `supabase_rebase_branch`
- `supabase_reset_branch`
- `supabase_run_notebook`
- `supabase_search_docs`
- `supabase_submit_feedback`
- `supabase_update_storage_config`

## 2. Migration result

- SQL source: `supabase/schema.sql`
- Tool used: `supabase_apply_migration` (name: `integra_initial_schema`)
- Result: `{"success":true}`

## 3. Tables present (public schema)

Verified with `supabase_list_tables` and `information_schema.tables`:

- `accounts`
- `batches`
- `disbursements`
- `finding_events`
- `findings`
- `students`

All six tables have RLS enabled.

## 4. Connection hosts

Project API URL (from `supabase_get_project_url`):
`https://bzguwmwapgodhayhwsfr.supabase.co`

Direct connection (port 5432):
`db.bzguwmwapgodhayhwsfr.supabase.co:5432`

Pooler connection (port 6543):
`aws-<n>-<region>.pooler.supabase.com:6543`
The MCP tools do not expose the pooler region or instance number, so the exact
pooler hostname cannot be resolved through MCP. It must be read from the
Supabase dashboard (Project Settings -> Database -> Connection pooling).

## 5. Password retrievability

The database password is NOT retrievable through the MCP. No MCP tool exposes
credentials or secrets. The database was reached through the MCP server's own
privileged connection (`user=postgres`, PostgreSQL 17.11), not with a
user-supplied password. The password must come from the Supabase dashboard or
the project's secret store. Do not store it in this file.
