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

Commands:
  help [<command>...]
    Show help.

  db-init
    Init db.

  db-migrate
    STUB: Migrate db.

  create --user=USER --password=PASSWORD
    Create user.

  delete --user=USER
    Delete user.

  revoke --user=USER
    Revoke user.

  restore --user=USER
    Restore user.

  list [<flags>]
    List active users.
    
    flags:
      --all  Show all users include revoked and delete

  auth --user=USER --password=PASSWORD
    Auth user.

  change-password --user=USER --password=PASSWORD
    Change password
```
