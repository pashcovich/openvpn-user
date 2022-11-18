# openvpn-user

## Disclaimer
```
- Not tested in production environments! 
```


Use it on your own risk =)

### Description 
A simple tool to use with openvpn when you need to use `–auth-user-pass-verify` or wherever you want

### Example
make sure `openvpn-user` binary available through `PATH` variable and you have [auth.sh](https://github.com/pashcovich/openvpn-user/blob/master/auth.sh) or [auth_totp.sh](https://github.com/pashcovich/openvpn-user/blob/master/auth_totp.sh) script with `+x` rights available to openvpn server

i.e. put binary to `/usr/local/sbin/` and auth script to `/etc/openvpn/scripts/` dir

part of openvpn server config
```
script-security 2
auth-user-pass-verify /etc/openvpn/scripts/auth.sh via-file
```


### Usage
```
usage: openvpn-user [<flags>] <command> [<args> ...]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --db.path="./openvpn-user.db"  path do openvpn-user db

Commands:
  help [<command>...]
    Show help.


  db-init
    Init db.


  db-migrate
    STUB: Migrate db.


  create --user=USER --password=PASSWORD
    Create user.

    --user=USER          Username.
    --password=PASSWORD  Password.

  delete --user=USER [<flags>]
    Delete user.

    -f, --force      delete from db.
    -u, --user=USER  Username.

  revoke --user=USER
    Revoke user.

    -u, --user=USER  Username.

  restore --user=USER
    Restore user.

    -u, --user=USER  Username.

  list [<flags>]
    List active users.

    -a, --all  Show all users include revoked and deleted.

  check --user=USER
    check user existent.

    -u, --user=USER  Username.

  auth --user=USER [<flags>]
    Auth user.

    -u, --user=USER          Username.
    -p, --password=PASSWORD  Password.
    -t, --totp=TOTP          TOTP code.

  change-password --user=USER --password=PASSWORD
    Change password

    -u, --user=USER          Username.
    -p, --password=PASSWORD  Password.

  update-secret --user=USER [<flags>]
    update OTP secret

    -u, --user=USER          Username.
    -s, --secret="generate"  Secret.

  register-app --user=USER --totp=TOTP
    register 2FA application

    -u, --user=USER  Username.
    -t, --totp=TOTP  TOTP.

  check-app --user=USER
    check 2FA application

    -u, --user=USER  Username.

  get-secret --user=USER
    get OTP secret

    -u, --user=USER  Username.

```
