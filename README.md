# openvpn-user

## Disclaimer
```diff
- Not tested in production environments! 
```


Use it on your own risk =)

### Description 
A simple tool to use with openvpn when you need to use `–auth-user-pass-verify` or wherever you want

### Example
make sure `openvpn-user` binary available through `PATH` variable and you have `auth.sh` script with `+x` rights available to openvpn server

i.e. put binary to `/usr/local/sbin/` and auth script to `/etc/openvpn/scripts/` dir

part of openvpn server config
```bash
script-security 2
auth-user-pass-verify /etc/openvpn/scripts/auth.sh via-file
```


### Usage
```
usage: openvpn-user [<flags>] <command> [<args> ...]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --db.path="./openvpn-user.db"  path do openvpn-user db
  --debug                        Enable debug mode.
  --version                      Show application version.

Commands:
  help [<command>...]
    Show help.

  db-init
    Init db.

  db-migrate
    STUB: Migrate db.

  create --user=USER --password=PASSWORD
    Create user.

  delete --user=USER [<flags>]
    Delete user.

  revoke --user=USER
    Revoke user.

  restore --user=USER
    Restore user.

  list [<flags>]
    List active users.

  check --user=USER
    check user existent.

  auth --user=USER [<flags>]
    Auth user.

  change-password --user=USER --password=PASSWORD
    Change password

  update-secret --user=USER [<flags>]
    update OTP secret

  register-app --user=USER
    register 2FA application

  get-secret --user=USER
    get OTP secret

```
