-- +goose Up
CREATE TABLE contracts(
    id text PRIMARY KEY,
    title text NOT NULL,
    active boolean NOT NULL DEFAULT FALSE,
    responsible text
);

CREATE TABLE events(
    id text PRIMARY KEY,
    contract_id text NOT NULL, -- no reference as it might be used for contract delete events
    "type" text NOT NULL,
    ts timestamptz NOT NULL DEFAULT now(),
    processed boolean NOT NULL DEFAULT FALSE,
    data jsonb NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION notify_listener()
    RETURNS TRIGGER
    AS $$
BEGIN
    PERFORM
        pg_notify('ev', 'ev');
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- +goose StatementEnd
CREATE OR REPLACE TRIGGER "after_insert_notify"
    AFTER INSERT ON events
    FOR EACH STATEMENT
    EXECUTE PROCEDURE notify_listener();

-- +goose Down
DROP TABLE contracts;

DROP TABLE events;

