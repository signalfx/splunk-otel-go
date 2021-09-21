# Splunk Instrumentation for the MySQL Driver Package

This package instruments the
[`github.com/go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql)
package using the [`splunksql`](../../../../database/sql/splunksql) package.

## Getting Started

This package is design to be a drop-in replacement for the existing use of the
`mysql` package. The blank identified imports of that package can be replaced
with this package, and the standard library `sql.Open` function can be replaced
with the equivalent `Open` from `splunksql`.

```golang
import (
	"time"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
    // Make sure this is imported to ensure driver is registered.
	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
)

func main() {
	db, err := splunksql.Open("mysql", "user:password@/dbname")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
    /* ... */
}
```
