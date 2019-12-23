package store

const CreateTables string = `
	CREATE TABLE IF NOT EXISTS "users" (
		"id"	INTEGER NOT NULL UNIQUE,
		"balance"	REAL NOT NULL DEFAULT 0,
		PRIMARY KEY("id")
	);
	CREATE TABLE IF NOT EXISTS "deposits" (
		"id"	INTEGER NOT NULL UNIQUE,
		"userId"	INTEGER NOT NULL,
		"balanceBefore"	REAL NOT NULL,
		"balanceAfter"	REAL NOT NULL,
		"date" INTEGER NOT NULL,
		PRIMARY KEY("id")
	);
	CREATE TABLE IF NOT EXISTS "transactions" (
		"id"	INTEGER NOT NULL UNIQUE,
		"userId"	INTEGER NOT NULL,
		"type"	TEXT NOT NULL,
		"amount"	REAL NOT NULL,
		"balanceBefore"	REAL NOT NULL,
		"balanceAfter"	REAL NOT NULL,
		"date" INTEGER NOT NULL,
		PRIMARY KEY("id")
	);
	`
const CreateIndexes string = `
	CREATE INDEX IF NOT EXISTS "transactionUserId" ON "transactions" ( "userId" ASC );
	CREATE INDEX IF NOT EXISTS "depositUserId" ON "deposits" ( "userId" ASC );
`
