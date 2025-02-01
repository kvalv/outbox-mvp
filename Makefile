db:
	PGPASSWORD=postgres psql -h localhost -U postgres -d postgres -p 5432
migrate:
	goose up
