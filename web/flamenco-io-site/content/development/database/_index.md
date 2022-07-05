---
title: Database
weight: 50
---

Flamenco Manager and Worker use SQLite as database, and Gorm as
object-relational mapper.

Since SQLite has limited support for altering table schemas, migration requires
copying old data to a temporary table with the new schema, then swap out the
tables. Because of this, avoid `NOT NULL` columns, as they will be problematic
in this process.
