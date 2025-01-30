CREATE TABLE employees(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL
);

CREATE TABLE contracts(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    title text NOT NULL,
    description text NULL,
    start_date timestamp NOT NULL,
    active boolean NOT NULL DEFAULT TRUE,
    responsible uuid REFERENCES employees(id)
);

CREATE TABLE contract_events(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id text NOT NULL, -- no reference as it might be used for contract delete events
    "type" text NOT NULL,
    ts timestamp NOT NULL GENERATED ALWAYS AS (now()) STORED,
    processed boolean NOT NULL DEFAULT FALSE,
    data jsonb NOT NULL
);

