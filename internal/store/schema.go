package store

const CreateTables string = `
	CREATE TABLE IF NOT EXISTS "users" (
		"id"	INTEGER NOT NULL UNIQUE,
		"balance"	REAL NOT NULL DEFAULT 0,
		PRIMARY KEY("id")
	);
	CREATE TABLE IF NOT EXISTS "deposits" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"balanceBefore"	REAL NOT NULL,
		"balanceAfter"	REAL NOT NULL,
		PRIMARY KEY("id")
	);
	CREATE TABLE IF NOT EXISTS "transactions" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"type"	TEXT NOT NULL,
		"amount"	REAL NOT NULL,
		PRIMARY KEY("id")
	);
	`
const CreateIndexes string = `
	CREATE INDEX IF NOT EXISTS "transaction_user_id" ON "transactions" ( "user_id" ASC );
	CREATE INDEX IF NOT EXISTS "deposit_user_id" ON "deposits" ( "user_id" ASC );
`
