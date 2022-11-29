package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pashcovich/openvpn-user/src"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

const (
	version = "1.0.7"
)

var (
	dbPath = kingpin.Flag("db.path", "path do openvpn-user db").Default("./openvpn-user.db").String()

	dbInitCommand    = kingpin.Command("db-init", "Init db.")
	dbMigrateCommand = kingpin.Command("db-migrate", "STUB: Migrate db.")

	createCommand             = kingpin.Command("create", "Create user.")
	createCommandUserFlag     = createCommand.Flag("user", "Username.").Required().String()
	createCommandPasswordFlag = createCommand.Flag("password", "Password.").Required().String()

	deleteCommand              = kingpin.Command("delete", "Delete user.")
	deleteCommandUserForceFlag = deleteCommand.Flag("force", "delete from db.").Short('f').Default("false").Bool()
	deleteCommandUserFlag      = deleteCommand.Flag("user", "Username.").Short('u').Required().String()

	revokeCommand         = kingpin.Command("revoke", "Revoke user.")
	revokeCommandUserFlag = revokeCommand.Flag("user", "Username.").Short('u').Required().String()

	restoreCommand         = kingpin.Command("restore", "Restore user.")
	restoreCommandUserFlag = restoreCommand.Flag("user", "Username.").Short('u').Required().String()

	listCommand        = kingpin.Command("list", "List active users.")
	listCommandAllFlag = listCommand.Flag("all", "Show all users include revoked and deleted.").Short('a').Default("false").Bool()

	checkCommand         = kingpin.Command("check", "check user existent.")
	checkCommandUserFlag = checkCommand.Flag("user", "Username.").Short('u').Required().String()

	authCommand             = kingpin.Command("auth", "Auth user.")
	authCommandUserFlag     = authCommand.Flag("user", "Username.").Short('u').Required().String()
	authCommandPasswordFlag = authCommand.Flag("password", "Password.").Short('p').String()
	authCommandTotpFlag     = authCommand.Flag("totp", "TOTP code.").Short('t').String()

	changePasswordCommand             = kingpin.Command("change-password", "Change password")
	changePasswordCommandUserFlag     = changePasswordCommand.Flag("user", "Username.").Short('u').Required().String()
	changePasswordCommandPasswordFlag = changePasswordCommand.Flag("password", "Password.").Short('p').Required().String()

	updateSecretCommand           = kingpin.Command("update-secret", "update OTP secret")
	updateSecretCommandUserFlag   = updateSecretCommand.Flag("user", "Username.").Short('u').Required().String()
	updateSecretCommandSecretFlag = updateSecretCommand.Flag("secret", "Secret.").Short('s').Default("generate").String()

	registerAppCommand         = kingpin.Command("register-app", "register 2FA application")
	registerAppCommandUserFlag = registerAppCommand.Flag("user", "Username.").Short('u').Required().String()
	registerAppCommandTotpFlag = registerAppCommand.Flag("totp", "TOTP.").Short('t').Required().String()

	resetAppCommand         = kingpin.Command("reset-app", "register 2FA application")
	resetAppCommandUserFlag = resetAppCommand.Flag("user", "Username.").Short('u').Required().String()

	checkAppCommand         = kingpin.Command("check-app", "check 2FA application")
	checkAppCommandUserFlag = checkAppCommand.Flag("user", "Username.").Short('u').Required().String()

	getSecretCommand         = kingpin.Command("get-secret", "get OTP secret")
	getSecretCommandUserFlag = getSecretCommand.Flag("user", "Username.").Short('u').Required().String()
)

func main() {

	args := kingpin.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		kingpin.Fatalf(err.Error())
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			kingpin.Fatalf(err.Error())
		}
	}(db)

	openvpnUser := src.OpenvpnUser{Database: db}

	kingpin.Version(version).VersionFlag.Short('v')

	switch args {
	case createCommand.FullCommand():
		wrap(openvpnUser.CreateUser(*createCommandUserFlag, *createCommandPasswordFlag))
	case deleteCommand.FullCommand():
		wrap(openvpnUser.DeleteUser(*deleteCommandUserFlag, *deleteCommandUserForceFlag))
	case revokeCommand.FullCommand():
		wrap(openvpnUser.RevokedUser(*revokeCommandUserFlag))
	case restoreCommand.FullCommand():
		wrap(openvpnUser.RestoreUser(*restoreCommandUserFlag))
	case listCommand.FullCommand():
		openvpnUser.PrintUsers(*listCommandAllFlag)
	case checkCommand.FullCommand():
		_ = openvpnUser.CheckUserExistent(*checkCommandUserFlag)
	case authCommand.FullCommand():
		provideAuthType := 0
		if *authCommandPasswordFlag != "" {
			provideAuthType += 1
		}
		if *authCommandTotpFlag != "" {
			provideAuthType += 1
		}
		if provideAuthType == 1 {
			authSuccessful, authErr := openvpnUser.AuthUser(*authCommandUserFlag, *authCommandPasswordFlag, *authCommandTotpFlag)
			if authErr != nil {
				kingpin.Fatalf(authErr.Error())
			} else if authSuccessful {
				fmt.Println("Authorization successful")
			}
		} else {
			fmt.Println("Please provide only one type of auth flag")
			os.Exit(1)
		}
	case changePasswordCommand.FullCommand():
		wrap(openvpnUser.ChangeUserPassword(*changePasswordCommandUserFlag, *changePasswordCommandPasswordFlag))
	case updateSecretCommand.FullCommand():
		wrap(openvpnUser.RegisterOtpSecret(*updateSecretCommandUserFlag, *updateSecretCommandSecretFlag))
	case registerAppCommand.FullCommand():
		wrap(openvpnUser.RegisterOtpApplication(*registerAppCommandUserFlag, *registerAppCommandTotpFlag))
	case resetAppCommand.FullCommand():
		wrap(openvpnUser.ResetOtpApplication(*resetAppCommandUserFlag))
	case checkAppCommand.FullCommand():
		appConfigured, appErr := openvpnUser.IsSecondFactorEnabled(*checkAppCommandUserFlag)
		if appErr != nil {
			kingpin.Fatalf(appErr.Error())
		} else if appConfigured {
			fmt.Println("App configured")
		}
	case getSecretCommand.FullCommand():
		wrap(openvpnUser.GetUserOtpSecret(*getSecretCommandUserFlag))

	case dbInitCommand.FullCommand():
		openvpnUser.InitDb()
	case dbMigrateCommand.FullCommand():
		openvpnUser.MigrateDb()
	}
}

func wrap(msg string, err error) {
	if err != nil {
		kingpin.Fatalf(err.Error())
	} else {
		fmt.Println(msg)
	}
}
