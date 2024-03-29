## Use AlloyDB as Go library

When you use AlloyDB as Go linary with a local storage engine, it works as an embedded database.
One of the benefits of having an embedded database is that you can have a completely isolated environment to run tests.
Removing the dependency of a real database for tests means you can run tests anywhere.

This is an example code to use AlloyDB as Go library.

```go
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/Dong-Chan/alloydb"
	"github.com/ngaut/log"

)

func main() {
	// Default log level is debug, set it to error to turn off debug log.
	log.SetLevelByString("error")

	// DriverName is 'alloydb', dataSourceName is in the form of "<engine>://<dbPath>".
	// dbPath is the directory that stores the data files if you use a local storage engine.
	dbPath := "/tmp/alloydb"
	db, err := sql.Open("alloydb", "goleveldb://" + dbPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS t (a INT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO t VALUES (?)", 1)
	if err != nil {
		log.Fatal(err)
	}

	row := db.QueryRow("SELECT * FROM t WHERE a = ?", 1)
	var val int
	err = row.Scan(&val)
	if err != nil {
		log.Fatal(err)
	}
	row.Close()

	fmt.Printf("value is %d\n", val)

	_, err = db.Exec("DROP TABLE t")
	if err != nil {
		log.Fatal(err)
	}
}

```
