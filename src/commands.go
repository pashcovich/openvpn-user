package src

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"text/tabwriter"
)

func (oUser *OpenvpnUser) InitDb() {
	// boolean fields are integer because of sqlite does not support boolean: 1 = true, 0 = false
	_, err := oUser.Database.Exec("CREATE TABLE IF NOT EXISTS users(id integer not null primary key autoincrement, username string UNIQUE, password string, secret string, revoked integer default 0, deleted integer default 0, app_configured integer default 0)")
	checkErr(err)
	_, err = oUser.Database.Exec("CREATE TABLE IF NOT EXISTS migrations(id integer not null primary key autoincrement, name string)")
	checkErr(err)
	fmt.Println("Database initialized")
}

func (oUser *OpenvpnUser) CreateUser(username, password string) (string, error) {
	if !oUser.CheckUserExistent(username) {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		_, err := oUser.Database.Exec("INSERT INTO users(username, password, secret, revoked, deleted, app_configured) VALUES ($1, $2, $3, 0, 0, 0)", username, string(hash), "")
		checkErr(err)
		return "User created", nil
	} else {
		return "", userAlreadyExistError

	}

}

func (oUser *OpenvpnUser) DeleteUser(username string, force bool) (string, error) {
	deleteQuery := "UPDATE users SET deleted = 1 WHERE username = $1"
	if force {
		deleteQuery = "DELETE FROM users WHERE username = $1"
	}
	res, err := oUser.Database.Exec(deleteQuery, username)
	if err != nil {
		return "", err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		return "", userDeleteError
	}
	return "User deleted", nil

}

func (oUser *OpenvpnUser) RevokedUser(username string) (string, error) {
	if !oUser.userDeleted(username) {
		res, err := oUser.Database.Exec("UPDATE users SET revoked = 1 WHERE username = $1", username)
		if err != nil {
			return "", err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return "", err
		}

		if rowsAffected == 0 {
			return "", userRevokeError
		}
		return "User revoked", nil
	}
	return "", userDeletedError
}

func (oUser *OpenvpnUser) RestoreUser(username string) (string, error) {
	if !oUser.userDeleted(username) {
		res, err := oUser.Database.Exec("UPDATE users SET revoked = 0 WHERE username = $1", username)
		if err != nil {
			return "", err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return "", err
		}

		if rowsAffected == 0 {
			return "", userRestoreError
		}
		return "User restored", nil
	}
	return "", userDeletedError
}

func (oUser *OpenvpnUser) CheckUserExistent(username string) bool {
	c := 0
	_ = oUser.Database.QueryRow("SELECT count(*) FROM users WHERE username = $1", username).Scan(&c)
	if c == 1 {
		return true
	} else {
		return false
	}
}

func (oUser *OpenvpnUser) userDeleted(username string) bool {
	u := User{}
	_ = oUser.Database.QueryRow("SELECT deleted FROM users WHERE username = $1", username).Scan(&u.deleted)
	if u.deleted {
		return true
	} else {
		return false
	}
}

func (oUser *OpenvpnUser) userIsActive(username string) bool {
	// return true if user exist and not deleted or revoked
	u := User{}
	err := oUser.Database.QueryRow("SELECT revoked,deleted FROM users WHERE username = $1", username).Scan(&u.revoked, &u.deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		return false
	}
	if !u.revoked && !u.deleted {
		return true
	} else {
		return false
	}
}

func (oUser *OpenvpnUser) listUsers(all bool) []User {
	var users []User
	condition := "WHERE deleted = 0 AND revoked = 0"

	if all {
		condition = ""
	}

	query := "SELECT id, username, password, revoked, deleted, app_configured FROM users " + condition

	rows, err := oUser.Database.Query(query)
	checkErr(err)

	for rows.Next() {
		u := User{}
		err = rows.Scan(&u.id, &u.name, &u.password, &u.revoked, &u.deleted, &u.appConfigured)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	return users
}

func (oUser *OpenvpnUser) PrintUsers(all bool) {
	ul := oUser.listUsers(all)
	if len(ul) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent|tabwriter.Debug)
		_, _ = fmt.Fprintln(w, "id\t username\t revoked\t deleted\t app_configured")
		for _, u := range ul {
			_, _ = fmt.Fprintf(w, "%d\t %s\t %v\t %v\t %v\n", u.id, u.name, u.revoked, u.deleted, u.appConfigured)
		}
		_ = w.Flush()
	} else {
		fmt.Println("No users created yet")
	}
}

func (oUser *OpenvpnUser) ChangeUserPassword(username, password string) (string, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	_, err := oUser.Database.Exec("UPDATE users SET password = $1 WHERE username = $2", hash, username)
	if err != nil {
		return "", err
	}

	return "Password changed", nil
}

func (oUser *OpenvpnUser) RegisterOtpSecret(username, secret string) (string, error) {
	if oUser.userIsActive(username) {
		if secret == "generate" {
			randomStr := RandStr(20, "num")

			secret = base32.StdEncoding.EncodeToString([]byte(randomStr))
		}

		_, err := oUser.Database.Exec("UPDATE users SET secret = $1 WHERE username = $2", secret, username)
		if err != nil {
			return "", err
		}

		return "Secret updated", nil
	}
	return "", userIsNotActiveError
}

func (oUser *OpenvpnUser) RegisterOtpApplication(username, totp string) (string, error) {
	if oUser.userIsActive(username) {

		appConfigured, appErr := oUser.IsSecondFactorEnabled(username)
		if appErr != nil {
			return "", appErr
		}
		if !appConfigured {

			authOk, authErr := oUser.AuthUser(username, "", totp)
			if authErr != nil {
				return "", authErr
			}
			if authOk {
				_, err := oUser.Database.Exec("UPDATE users SET app_configured = 1 WHERE username = $1", username)
				if err != nil {
					return "", err
				}
				return "OTP application configured", nil
			}
		}
		return "OTP application already configured", nil
	}
	return "", userIsNotActiveError
}
func (oUser *OpenvpnUser) ResetOtpApplication(username string) (string, error) {
	if oUser.userIsActive(username) {

		appConfigured, appErr := oUser.IsSecondFactorEnabled(username)
		if appErr != nil {
			return "", appErr
		}
		if appConfigured {
			_, err := oUser.Database.Exec("UPDATE users SET app_configured = 0 WHERE username = $1", username)
			if err != nil {
				return "", err
			}
			return "OTP application reset successful", nil
		}
		return "OTP application not configured", nil
	}
	return "", userIsNotActiveError
}

func (oUser *OpenvpnUser) GetUserOtpSecret(username string) (string, error) {
	if oUser.userIsActive(username) {
		u := User{}
		_ = oUser.Database.QueryRow("SELECT secret FROM users WHERE username = $1", username).Scan(&u.secret)

		return u.secret, nil
	}
	return "", userIsNotActiveError
}
func (oUser *OpenvpnUser) IsSecondFactorEnabled(username string) (bool, error) {
	if oUser.userIsActive(username) {
		u := User{}
		err := oUser.Database.QueryRow("SELECT username, app_configured FROM users WHERE username = $1", username).Scan(&u.name, &u.appConfigured)
		if err != nil {
			return false, err
		}

		if u.name == username {
			return u.appConfigured, nil
		}
		
		return false, checkAppError
	}
	return false, userIsNotActiveError
}

func (oUser *OpenvpnUser) AuthUser(username, password, totp string) (bool, error) {

	row := oUser.Database.QueryRow("SELECT id, username, password, revoked, deleted, secret, app_configured FROM users WHERE username = $1", username)
	u := User{}
	err := row.Scan(&u.id, &u.name, &u.password, &u.revoked, &u.deleted, &u.secret, &u.appConfigured)
	if err != nil {
		return false, err
	}

	if oUser.userIsActive(username) {
		if password == "" && len(totp) > 0 {
			if len(u.secret) == 0 {
				return false, userSecretDoesNotExistError
			}

			otpConfig := &dgoogauth.OTPConfig{
				Secret:      strings.TrimSpace(u.secret),
				WindowSize:  3,
				HotpCounter: 0,
			}

			trimmedToken := strings.TrimSpace(totp)

			ok, authErr := otpConfig.Authenticate(trimmedToken)

			if authErr != nil {
				fmt.Println(authErr)
			}
			if ok {
				return true, nil
			} else {
				return false, tokenMismatchedError
			}

		} else if len(password) > 0 && totp == "" {
			err = bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
			if err != nil {
				return false, passwordMismatchedError
			} else {
				return true, nil

			}
		}
	}
	return false, userIsNotActiveError

}

func (oUser *OpenvpnUser) MigrateDb() {
	var c int
	var migrations []Migration

	migrations = append(migrations, Migration{name: "users_add_secret_column_2022_11_10", sql: "ALTER TABLE users ADD COLUMN secret string"})
	migrations = append(migrations, Migration{name: "users_add_2fa_column_2022_11_11", sql: "ALTER TABLE users ADD COLUMN app_configured integer default 0"})

	for _, migration := range migrations {
		c = -1
		err := oUser.Database.QueryRow("SELECT count(*) FROM migrations WHERE name = $1", migration.name).Scan(&c)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			fmt.Println(err)
		}
		if c == 0 {
			fmt.Printf("Migrating database with new migration %s\n", migration.name)
			_, err = oUser.Database.Exec(migration.sql)
			checkErr(err)
			_, err = oUser.Database.Exec("INSERT INTO migrations(name) VALUES ($1)", migration.name)
			checkErr(err)
		}
	}
	fmt.Println("Migrations are up to date")
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
