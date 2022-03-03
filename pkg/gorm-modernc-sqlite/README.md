# GORM Sqlite Driver

This is a clone of the [Gorm SQLite driver](https://github.com/go-gorm/sqlite),
adjusted by Sybren Stüvel <sybren@blender.org> to use modernc.org/sqlite instead
of the SQLite C-bindings wrapper.


## USAGE

```go
import (
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
)

// modernc.org/sqlite
db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
```

Checkout [https://gorm.io](https://gorm.io) for details.
