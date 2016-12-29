package common_test

import (
	. "common"
	"fmt"
	"testing"
)

func TestDB(t *testing.T) {
	dbPool, err := NewDBPoolWithConfig("db.cfg", "local_db")

	if err == nil {
		db := dbPool.GetConn()
		sql := "select count(1) from tb1"

		rets := dbPool.FetchOneTuple(sql, db)

		fmt.Println(rets)

		dbPool.Release()
	}
}
