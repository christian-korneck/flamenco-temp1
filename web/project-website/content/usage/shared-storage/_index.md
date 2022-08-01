---
title: Shared Storage
---

Flamenco needs some form of *shared storage*: a place for files to be stored
that can be accessed by all the computers in the farm.

Basically there are three approaches to this:

| Approach                            | Simple | Efficient | Render jobs are isolated |
|-------------------------------------|--------|-----------|--------------------------|
| Work directly on the shared storage | ✅      | ✅         | ❌                        |
| Create a copy for each render job   | ✅      | ❌         | ✅                        |
| Shaman Storage System               | ❌      | ✅         | ✅                        |

Each is explained below.

## Work Directly on the Shared Storage

Working directly in the shared storage is the simplest way to work with Flamenco.

## Creating a Copy for Each Render Job

The "work on shared storage" approach has the downside that render jobs are not
fully separated from each other. For example, when you change a texture while a
render job is running, the subsequently rendered frames will be using that
altered texture. If this is an issue for you, and you cannot use the [Shaman
Storage System][shaman], the approach described in this section is for you.

[shaman]: #shaman-storage-system

## Shaman Storage System

TODO: write

## Platform-specific Notes

### Windows

The Shaman storage system uses _symbolic links_. On Windows the creation of
symbolic links requires a change in security policy. Unfortunately, *Home*
editions of Windows do not have a policy editor, but the freely available
[Polsedit][polsedit] can be used on these editions.

1. Press Win+R, in the popup type `secpol.msc`. Then click OK.
2. In the _Local Security Policy_ window that opens, go to _Security Settings_ > _Local Policies_ > _User Rights Assignment_.
3. In the list, find the _Create Symbolic Links_ item.
4. Double-click the item and add yourself (or the user running Flamenco Manager or the whole users group) to the list.
5. Log out & back in again, or reboot the machine.

[polsedit]: https://www.southsoftware.com/polsedit.html


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
