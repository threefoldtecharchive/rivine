# Migrations

A bunch of temporary (throw-away) migration tools,
to help developers migrate breaking changes.

## UnlockHash Checksum Fix

Issues that was fixed:

> In previous milestone (0.6) we introduced the change where the unlock type
> was prefixed to the unlock hash. In hex format this is 2 extra bytes.
>
> The checksum however was still only done using the the actual hash as input,
> meaning a typo in the hash type would not get detected by the checksum.
>
> Github issue: https://github.com/rivine/rivine/issues/219

How it was fixed:

> The checksum now gets generated using both the unlock type as well as the hash.

This tool helps you migrate old addresses (unlock hashes),
where the checksum is still only computed using the hash,
to the new addresses (unlock hashes),
where the checksum is computed using both the type and the hash.

Using it is as simple as:

```
$ go run cmd/migration/unlockhash_checksum_fix.go 01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec15c28ee7d7ed1d
01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e
```

If you run it given a unlock hash (address) already in the new format, you'll get an error:

```
$ go run cmd/migration/unlockhash_checksum_fix.go 01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e
2018/03/30 13:06:36 [Error] unlockhash_checksum_fix.go:23 Given unlock hash is already in the new correct format
exit status 1
```
