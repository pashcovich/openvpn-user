package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"text/tabwriter"
)

const (
	version = "1.0.4"
)

var (
	dbPath = kingpin.Flag("db.path", "path do openvpn-user db").Default("./openvpn-user.db").String()

	dbInitCommand    = kingpin.Command("db-init", "Init db.")
	dbMigrateCommand = kingpin.Command("db-migrate", "STUB: Migrate db.")

	createCommand             = kingpin.Command("create", "Create user.")
	createCommandUserFlag     = createCommand.Flag("user", "Username.").Required().String()
	createCommandPasswordFlag = createCommand.Flag("password", "Password.").Required().String()

	deleteCommand              = kingpin.Command("delete", "Delete user.")
	deleteCommandUserForceFlag = deleteCommand.Flag("force", "delete from db.").Default("false").Bool()
	deleteCommandUserFlag      = deleteCommand.Flag("user", "Username.").Required().String()

	revokeCommand         = kingpin.Command("revoke", "Revoke user.")
	revokeCommandUserFlag = revokeCommand.Flag("user", "Username.").Required().String()

	restoreCommand         = kingpin.Command("restore", "Restore user.")
	restoreCommandUserFlag = restoreCommand.Flag("user", "Username.").Required().String()

	listCommand = kingpin.Command("list", "List active users.")
	listAll     = listCommand.Flag("all", "Show all users include revoked and deleted.").Default("false").Bool()

	checkCommand         = kingpin.Command("check", "check user existent.")
	checkCommandUserFlag = checkCommand.Flag("user", "Username.").Required().String()

	authCommand             = kingpin.Command("auth", "Auth user.")
	authCommandUserFlag     = authCommand.Flag("user", "Username.").Required().String()
	authCommandPasswordFlag = authCommand.Flag("password", "Password.").Required().String()

	changePasswordCommand             = kingpin.Command("change-password", "Change password")
	changePasswordCommandUserFlag     = changePasswordCommand.Flag("user", "Username.").Required().String()
	changePasswordCommandPasswordFlag = changePasswordCommand.Flag("password", "Password.").Required().String()

	debug = kingpin.Flag("debug", "Enable debug mode.").Default("false").Bool()
)

type User struct {
	id       int64
	name     string
	password string
	revoked  bool
	deleted  bool
}

func main() {
	kingpin.Version(version)
	switch kingpin.Parse() {
	case createCommand.FullCommand():
		createUser(*createCommandUserFlag, *createCommandPasswordFlag)
	case deleteCommand.FullCommand():
		deleteUser(*deleteCommandUserFlag)
	case revokeCommand.FullCommand():
		revokedUser(*revokeCommandUserFlag)
	case restoreCommand.FullCommand():
		restoreUser(*restoreCommandUserFlag)
	case listCommand.FullCommand():
		printUsers()
	case checkCommand.FullCommand():
		_ = checkUserExistent(*checkCommandUserFlag)
	case authCommand.FullCommand():
		authUser(*authCommandUserFlag, *authCommandPasswordFlag)
	case changePasswordCommand.FullCommand():
		changeUserPassword(*changePasswordCommandUserFlag, *changePasswordCommandPasswordFlag)
	case dbInitCommand.FullCommand():
		initDb()
	case dbMigrateCommand.FullCommand():
		migrateDb()
	}
}

func getDb() *sql.DB {
	db, err := sql.Open("sqlite3", *dbPath)
	checkErr(err)
	if db == nil {
		panic("db is nil")
	}
	return db
}

func initDb() {
	// boolean fields are integer because of sqlite does not support boolean: 1 = true, 0 = false
	_, err := getDb().Exec("CREATE TABLE IF NOT EXISTS users(id integer not null primary key autoincrement, username string UNIQUE, password string, revoked integer default 0, deleted integer default 0)")
	checkErr(err)
	fmt.Printf("Database initialized at %s\n", *dbPath)
}

func migrateDb() {
	fmt.Println("STUB: Migrations are up to date")
}

func createUser(username, password string) {
	if !checkUserExistent(username) {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		_, err := getDb().Exec("INSERT INTO users(username, password) VALUES ($1, $2)", username, string(hash))
		checkErr(err)
		fmt.Printf("User %s created\n", username)
	} else {
		fmt.Printf("ERROR: User %s already registered\n", username)
		os.Exit(1)
	}

}

func deleteUser(username string) {
	deleteQuery := "UPDATE users SET deleted = 1 WHERE username = $1"
	if *deleteCommandUserForceFlag {
		deleteQuery = "DELETE FROM users WHERE username = $1"
	}
	res, err := getDb().Exec(deleteQuery, username)
	checkErr(err)
	if rowsAffected, rowsErr := res.RowsAffected(); rowsErr != nil {
		if rowsAffected == 1 {
			fmt.Printf("User %s deleted\n", username)
		}
	} else {
		if *debug {
			fmt.Printf("ERROR: due deleting user %s: %s\n", username, rowsErr)
		}
	}
}

func revokedUser(username string) {
	if !userDeleted(username) {
		res, err := getDb().Exec("UPDATE users SET revoked = 1 WHERE username = $1", username)
		checkErr(err)
		if rowsAffected, rowsErr := res.RowsAffected(); rowsErr != nil {
			if rowsAffected == 1 {
				fmt.Printf("User %s revoked\n", username)
			}
		} else {
			if *debug {
				fmt.Printf("ERROR: due reoking user %s: %s\n", username, rowsErr)
			}
		}
	}
}

func restoreUser(username string) {
	if !userDeleted(username) {
		res, err := getDb().Exec("UPDATE users SET revoked = 0 WHERE username = $1", username)
		checkErr(err)
		if rowsAffected, rowsErr := res.RowsAffected(); rowsErr != nil {
			if rowsAffected == 1 {
				fmt.Printf("User %s restored\n", username)
			}
		} else {
			if *debug {
				fmt.Printf("ERROR: due restoring user %s: %s\n", username, rowsErr)
			}
		}
	}
}

func checkUserExistent(username string) bool {
	// we need to check if there is already such a user
	// return true if user exist
	var c int
	_ = getDb().QueryRow("SELECT count(*) FROM users WHERE username = $1", username).Scan(&c)
	if c == 1 {
		fmt.Printf("User %s exist\n", username)
		return true
	} else {
		return false
	}
}

func userDeleted(username string) bool {
	// return true if user marked as deleted
	u := User{}
	_ = getDb().QueryRow("SELECT * FROM users WHERE username = $1", username).Scan(&u)
	if u.deleted {
		fmt.Printf("User %s marked as deleted\n", username)
		return true
	} else {
		return false
	}
}

func userIsActive(username string) bool {
	// return true if user exist and not deleted and revoked
	u := User{}
	_ = getDb().QueryRow("SELECT * FROM users WHERE username = $1", username).Scan(&u)
	if !u.revoked && !u.deleted {
		fmt.Printf("User %s is active\n", username)
		return true
	} else {
		fmt.Println("User may be deleted or revoked")
		return false
	}
}

func listUsers() []User {
	condition := "WHERE deleted = 0 AND revoked = 0"
	var users []User
	if *listAll {
		condition = ""
	}
	query := "SELECT * FROM users " + condition
	rows, err := getDb().Query(query)
	checkErr(err)

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.id, &u.name, &u.password, &u.revoked, &u.deleted)
		if err != nil {
			fmt.Println(err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func printUsers() {
	ul := listUsers()
	if len(ul) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent|tabwriter.Debug)
		_, _ = fmt.Fprintln(w, "id\t username\t revoked\t deleted")
		for _, u := range ul {
			fmt.Fprintf(w, "%d\t %s\t %v\t %v\n", u.id, u.name, u.revoked, u.deleted)
		}
		_ = w.Flush()
	} else {
		fmt.Println("No users created yet")
	}
}

func changeUserPassword(username, password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	_, err := getDb().Exec("UPDATE users SET password = $1 WHERE username = $2", hash, username)
	checkErr(err)

	fmt.Println("Password changed")
}

func authUser(username, password string) {

	row := getDb().QueryRow("select * from users where username = $1", username)
	u := User{}
	err := row.Scan(&u.id, &u.name, &u.password, &u.revoked, &u.deleted)
	checkErr(err)

	if userIsActive(username) {
		err = bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
		if err != nil {
			fmt.Println("Authorization failed")
			if *debug {
				fmt.Println("Passwords mismatched")
			}
			os.Exit(1)
		} else {
			fmt.Println("Authorization successful")
			os.Exit(0)
		}
	}
	fmt.Println("Authorization failed")
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
