package src

import "database/sql"

type OpenvpnUser struct {
	Database *sql.DB
}

type Migration struct {
	id   int64
	name string
	sql  string
}

type User struct {
	id            int64
	name          string
	password      string
	revoked       bool
	deleted       bool
	secret        string
	appConfigured bool
}
