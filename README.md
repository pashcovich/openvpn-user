# openvpn-user

### disclaimer
Not tested in production environments!

Use it on your own risk =)

### Description 
A simple tool to use with openvpn when you need to use `–auth-user-pass-verify` or wherever you want

### Example

part of openvpn server config
```bash
auth-user-pass-verify /etc/openvpn/scripts/auth.sh via-file
```

make sure `openvpn-user` binary available through `PATH` variable 
i.e. put it in `/usr/local/sbin/openvpn-user`

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
