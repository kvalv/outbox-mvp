// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package queries

import (
	"context"
)

const activateContract = `-- name: ActivateContract :one
UPDATE
    contracts
SET
    active = TRUE
WHERE
    id = $1
RETURNING
    id, title, active, responsible
`

func (q *Queries) ActivateContract(ctx context.Context, id string) (Contract, error) {
	row := q.db.QueryRow(ctx, activateContract, id)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Active,
		&i.Responsible,
	)
	return i, err
}

const assignResponsible = `-- name: AssignResponsible :exec
UPDATE
    contracts
SET
    responsible = $2
WHERE
    id = $1
`

type AssignResponsibleParams struct {
	ID          string
	Responsible string
}

func (q *Queries) AssignResponsible(ctx context.Context, arg AssignResponsibleParams) error {
	_, err := q.db.Exec(ctx, assignResponsible, arg.ID, arg.Responsible)
	return err
}

const createContract = `-- name: CreateContract :one
INSERT INTO contracts(id, title)
    VALUES ($1, $2)
RETURNING
    id, title, active, responsible
`

type CreateContractParams struct {
	ID    string
	Title string
}

func (q *Queries) CreateContract(ctx context.Context, arg CreateContractParams) (Contract, error) {
	row := q.db.QueryRow(ctx, createContract, arg.ID, arg.Title)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Active,
		&i.Responsible,
	)
	return i, err
}

const getContract = `-- name: GetContract :one
SELECT
    id, title, active, responsible
FROM
    contracts
WHERE
    id = $1
`

func (q *Queries) GetContract(ctx context.Context, id string) (Contract, error) {
	row := q.db.QueryRow(ctx, getContract, id)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Active,
		&i.Responsible,
	)
	return i, err
}

const getUnsentEvents = `-- name: GetUnsentEvents :many
SELECT
    id, contract_id, type, ts, processed, data
FROM
    events
WHERE
    processed = FALSE
FOR UPDATE
    SKIP LOCKED
`

func (q *Queries) GetUnsentEvents(ctx context.Context) ([]Event, error) {
	rows, err := q.db.Query(ctx, getUnsentEvents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.ContractID,
			&i.Type,
			&i.Ts,
			&i.Processed,
			&i.Data,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertEvent = `-- name: InsertEvent :exec
INSERT INTO events(id, contract_id, "type", data)
    VALUES ($1, $2, $3, $4)
`

type InsertEventParams struct {
	ID         string
	ContractID string
	Type       string
	Data       []byte
}

func (q *Queries) InsertEvent(ctx context.Context, arg InsertEventParams) error {
	_, err := q.db.Exec(ctx, insertEvent,
		arg.ID,
		arg.ContractID,
		arg.Type,
		arg.Data,
	)
	return err
}

const markEventAsProcessed = `-- name: MarkEventAsProcessed :one
UPDATE
    events
SET
    processed = TRUE
WHERE
    id = $1
RETURNING
    id, contract_id, type, ts, processed, data
`

func (q *Queries) MarkEventAsProcessed(ctx context.Context, id string) (Event, error) {
	row := q.db.QueryRow(ctx, markEventAsProcessed, id)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.ContractID,
		&i.Type,
		&i.Ts,
		&i.Processed,
		&i.Data,
	)
	return i, err
}

const notifyEvent = `-- name: NotifyEvent :exec
SELECT
    pg_notify('events', $1)
`

func (q *Queries) NotifyEvent(ctx context.Context, pgNotify string) error {
	_, err := q.db.Exec(ctx, notifyEvent, pgNotify)
	return err
}
