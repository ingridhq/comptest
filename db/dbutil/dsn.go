package dbutil

import "strings"

// SplitDSN will extract DSN (without db name) and database name.
func SplitDSN(dsn string) (string, string) {
	tmp := strings.Split(dsn, "/")
	newDSN := strings.Join(tmp[:len(tmp)-1], "/")
	dbName := strings.Split(tmp[len(tmp)-1], "?")[0]
	return newDSN, dbName
}
