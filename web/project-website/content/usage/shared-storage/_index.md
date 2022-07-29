---
title: Shared Storage
---

Flamenco needs some form of *shared storage*: a place for files to be stored
that can be accessed by all the computers in the farm.

TODO: write more about this.

## Shaman Storage System

TODO: write

## Platform-specific Notes

### Windows

TODO: lots of tricky stuff about getting symlinks to work on Windows. We may want to get inspiration from the [Git-for-Windows](https://github.com/git-for-windows/git/wiki/Symbolic-Links#allowing-non-administrators-to-create-symbolic-links) documentation.

### Linux

For symlinks to work with CIFS/Samba filesystems (like a typical NAS), you need
to mount it with the option `mfsymlinks`. As a concrete example, for a user
`sybren`, put something like this in `fstab`:

```
//NAS/flamenco /media/flamenco cifs mfsymlinks,credentials=/home/sybren/.smbcredentials,uid=sybren,gid=users 0 0
```

Then put the NAS credentials in `/home/sybren/.smbcredentials`:

```
username=sybren
password=g1mm3acce55plz
```

and be sure to protect it with `chmod 600 /home/sybren/.smbcredentials`.

Finally `mkdir /media/flamenco` and `sudo mount /media/flamenco` should get things mounted.

The above info was obtained from [Ask Ubuntu](https://askubuntu.com/a/157140).
