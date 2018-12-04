# checkers

The `health` library comes with a number of built-in checkers for well known
types of dependencies.

If a pre-built checker is not available, you can create your own checkers by
implementing the `ICheckable` interface (which consists of a single method -
`Status() (interface{}, error)`).

If you do create a custom-checker - consider opening a PR and adding it to the
list of built-in checkers.

## Built-in checkers

- [HTTP](#http)
- [Redis](#redis)
- [SQL DB](#sql-db)
- [Mongo](#mongo)
- [Reachable](#reachable)

### HTTP

The HTTP checker is a generic HTTP call executor. To make use of it, instantiate and fill out a `HTTPConfig` struct and pass it into `checkers.NewHTTP(...)`.

The only **required** attribute is `HTTPConfig.URL` (`*url.URL`).
Refer to the source code for all available attributes on the struct.

### Redis

The Redis checker allows you to test that your server is either available (by ping), is able to set a value, is able to get a value or all of the above.

To make use of it, instantiate and fill out a `RedisConfig` struct and pass it to `checkers.NewRedis(...)`.

The `RedisConfig` must contain a valid `RedisAuthConfig` and at least _one_ check method (ping, set or get).

Refer to the godocs for additional info.

### SQL DB

The SQL DB checker has implementations for the following interfaces:

- `SQLPinger`, which encloses `PingContext` in [`sql.DB`](https://golang.org/pkg/database/sql/#DB.PingContext) and [`sql.Conn`](https://golang.org/pkg/database/sql/#Conn.PingContext)
- `SQLQueryer`, which encloses `QueryContext` in [`sql.DB`](https://golang.org/pkg/database/sql/#DB.QueryContext), [`sql.Conn`](https://golang.org/pkg/database/sql/#Conn.QueryContext), [`sql.Stmt`](https://golang.org/pkg/database/sql/#Stmt.QueryContext), and [`sql.Tx`](https://golang.org/pkg/database/sql/#Tx.QueryContext)
- `SQLExecer`, which encloses `ExecContext` in [`sql.DB`](https://golang.org/pkg/database/sql/#DB.ExecContext), [`sql.Conn`](https://golang.org/pkg/database/sql/#Conn.ExecContext), [`sql.Stmt`](https://golang.org/pkg/database/sql/#Stmt.ExecContext), and [`sql.Tx`](https://golang.org/pkg/database/sql/#Tx.ExecContext)

#### SQLConfig
The `SQLConfig` struct is required when using the SQL DB health check.  It **must** contain an inplementation of one of either `SQLPinger`, `SQLQueryer`, or `SQLExecer`.

If `SQLQueryer` or `SQLExecer` are implemented, then `Query` must be valid (len > 0).

Additionally, if `SQLQueryer` or `SQLExecer` are implemented, you have the option to also set either the `QueryerResultHandler` or `ExecerResultHandler` functions.  These functions allow you to evaluate the result of a query or exec operation.  If you choose not to implement these yourself, the default handlers are used.

The default `ExecerResultHandler` is successful if the passed exec operation affected one and only one row.

The default `QueryerResultHandler` is successful if the passed query operation returned one and only one row.

#### SQLPinger
Use the `SQLPinger` interface if your health check is only concerned with your application's database connectivity. All you need to do is set the `Pinger` value in your `SQLConfig`.

```golang
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
		Pinger: db
	})
	if err != nil {
		return err
	}

	hc := health.New()
	healthCheck.AddCheck(&health.Config{
		Name:     "sql-check",
		Checker:  sqlCheck,
		Interval: time.Duration(3) * time.Second,
		Fatal:    true,
	})
```

#### SQLQueryer
Use the `SQLQueryer` interface if your health check requires you to read rows from your database.  You can optionally supply a query result handler function.  If you don't supply one, the default function will be used.  The function signature for the handler is:

```golang
type SQLQueryerResultHandler func(rows *sql.Rows) (bool, error)
```
The default query handler returns true if there was exactly one row in the resultset:

```golang
	func DefaultQueryHandler(rows *sql.Rows) (bool, error) {
		defer rows.Close()

		numRows := 0
		for rows.Next() {
			numRows++
		}

		return numRows == 1, nil
	}
```

**IMPORTANT**: Note that your query handler is responsible for closing the passed `*sql.Rows` value.

Sample `SQLQueryer` implementation:

```golang
	// this is our custom query row handler
	func myQueryHandler(rows *sql.Rows) (bool, error) {
		defer rows.Close()

		var healthValue string
		for rows.Next() {
			// this query will ever return at most one row
			if err := rows.Scan(&healthValue); err != nil {
				return false, err
			}
		}

		return healthValue == "ok", nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// we pass the id we are looking for inside the params value
	sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
		Queryerer:            db,
		Query:                "SELECT healthValue FROM some_table WHERE id = ?",
		Params:               []interface{}{1},
		QueryerResultHandler: myQueryHandler
	})
	if err != nil {
		return err
	}

	hc := health.New()
	healthCheck.AddCheck(&health.Config{
		Name:     "sql-check",
		Checker:  sqlCheck,
		Interval: time.Duration(3) * time.Second,
		Fatal:    true,
	})
```

#### SQLExecer
Use the `SQLExecer` interface if your health check requires you to update or insert to your database.  You can optionally supply an exec result handler function.  If you don't supply one, the default function will be used.  The function signature for the handler is:

```golang
type SQLExecerResultHandler func(result sql.Result) (bool, error)
```

The default exec handler returns true if there was exactly one affected row:

```golang
	func DefaultExecHandler(result sql.Result) (bool, error) {
		affectedRows, err := result.RowsAffected()
		if err != nil {
			return false, err
		}

		return affectedRows == int64(1), nil
	}
```

Sample `SQLExecer ` implementation:

```golang
	// this is our custom exec result handler
	func myExecHandler(result sql.Result) (bool, error) {
		insertId, err := result.LastInsertId()
		if err != nil {
			return false, err
		}

		// for this example, a check isn't valid
		// until after the 100th iteration
		return insertId > int64(100), nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	sqlCheck, err := checkers.NewSQL(&checkers.SQLConfig{
		Execer:              db,
		Query:               "INSERT INTO checks (checkTS) VALUES (NOW())",
		ExecerResultHandler: myExecHandler
	})
	if err != nil {
		return err
	}

	hc := health.New()
	healthCheck.AddCheck(&health.Config{
		Name:     "sql-check",
		Checker:  sqlCheck,
		Interval: time.Duration(3) * time.Second,
		Fatal:    true,
	})
```

### Mongo

Mongo checker allows you to test if an instance of MongoDB is available by using the underlying driver's ping method or check whether a collection exists or not.

To make use of it, initialize a `MongoConfig` struct and pass it to `checkers.NewMongo(...)`.

The `MongoConfig` struct must specify either one or both of the `Ping` or `Collection` fields.

### Reachable

The reachable checker is a generic TCP/UDP checker. Use it to verify that a configured address can be contacted via a request over TCP or UDP. This is useful if you do not care about a response from the target and simply want to know if the URL is reachable.

The only **required** attribute is `ReachableConfig.URL` (`*url.URL`).
Refer to the source code for all available attributes on the struct.
