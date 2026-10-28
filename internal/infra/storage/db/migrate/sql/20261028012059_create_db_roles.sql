-- +goose Up
-- +goose StatementBegin
DO
$$
    DECLARE
        r_name text;
    BEGIN
        FOREACH r_name IN ARRAY ARRAY['showcase_writer', 'showcase_reader']
            LOOP
                IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = r_name) THEN
                    EXECUTE format('CREATE ROLE %I WITH LOGIN PASSWORD %L', r_name, 'password');
                END IF;
            END LOOP;

        EXECUTE 'GRANT CONNECT ON DATABASE ' || quote_ident(current_database()) || ' TO showcase_writer, showcase_reader';
    END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO showcase_writer, showcase_reader;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO showcase_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO showcase_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO showcase_reader;

-- +goose Down
-- +goose StatementBegin
DO
$$
    DECLARE
        role_name text;
    BEGIN
        FOREACH role_name IN ARRAY ARRAY ['showcase_writer', 'showcase_reader']
            LOOP
                IF EXISTS (SELECT FROM pg_roles WHERE rolname = role_name) THEN
                    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM %I', role_name);
                    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM %I', role_name);
                    EXECUTE format('REVOKE ALL ON ALL TABLES IN SCHEMA public FROM %I', role_name);
                    EXECUTE format('REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM %I', role_name);
                    EXECUTE format('REVOKE USAGE ON SCHEMA public FROM %I', role_name);
                    EXECUTE format('REVOKE CONNECT ON DATABASE %I FROM %I', current_database(), role_name);

                    IF NOT EXISTS (
                        SELECT 1
                        FROM pg_shdepend
                        WHERE refclassid = 'pg_authid'::regclass
                          AND refobjid = (SELECT oid FROM pg_roles WHERE rolname = role_name)
                    ) THEN
                        EXECUTE format('DROP ROLE %I', role_name);
                    END IF;
                END IF;
            END LOOP;
    END
$$;
-- +goose StatementEnd
