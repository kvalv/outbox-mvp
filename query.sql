-- name: CreateContract :one
INSERT INTO contracts(title, start_date)
    VALUES ($1, $2)
RETURNING
    *;

-- name: ActivateContract :one
UPDATE
    contracts
SET
    active = TRUE
WHERE
    id = $1
RETURNING
    *;

-- name: DeactivateContract :one
UPDATE
    contracts
SET
    active = FALSE
WHERE
    id = $1
RETURNING
    *;

-- name: GetContract :one
SELECT
    *
FROM
    contracts
WHERE
    id = $1;

-- name: DeleteContract :one
DELETE FROM contracts
WHERE id = $1
RETURNING
    *;

-- name: GetContractsForResponsible :many
SELECT
    *
FROM
    contracts
WHERE
    responsible = $1;

-- name: CreateContractEvent :one
INSERT INTO contract_events(contract_id, "type", data)
    VALUES ($1, $2, $3)
RETURNING
    *;


-- name: NotifyEvent :one
select pg_notify('events', $1);

-- name: GetUnsentEvents :many
SELECT
    *
FROM
    contract_events
WHERE
    processed = FALSE
    FOR UPDATE SKIP LOCKED;

-- name: MarkEventAsProcessed :one
UPDATE
    contract_events
SET
    processed = TRUE
WHERE
    id = $1
RETURNING
    *;
