-- name: CreateContract :one
INSERT INTO contracts(id, title)
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

-- name: GetContract :one
SELECT
    *
FROM
    contracts
WHERE
    id = $1;

-- name: AssignResponsible :exec
UPDATE
    contracts
SET
    responsible = $2
WHERE
    id = $1;

-- name: InsertEvent :exec
INSERT INTO events(id, contract_id, "type", data)
    VALUES ($1, $2, $3, $4)e

-- name: MarkEventAsProcessed :one
UPDATE
    events
SET
    processed = TRUE
WHERE
    id = $1
RETURNING
    *;

-- name: NotifyEvent :exec
SELECT
    pg_notify('events', $1);

-- name: GetUnsentEvents :many
SELECT
    *
FROM
    events
WHERE
    processed = FALSE
FOR UPDATE
    SKIP LOCKED;

