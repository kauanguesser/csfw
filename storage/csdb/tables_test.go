// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package csdb_test

import (
	"sort"
	"testing"

	"context"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/util/cstesting"
	"github.com/corestoreio/csfw/util/magento"
	"github.com/corestoreio/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTableService(t *testing.T) {
	t.Parallel()
	assert.Equal(t, csdb.MustNewTables().Len(), 0)

	const (
		TableIndexStore   = iota + 1 // Table: store
		TableIndexGroup              // Table: store_group
		TableIndexWebsite            // Table: store_website
		TableIndexZZZ                // the maximum index, which is not available.
	)

	tm1 := csdb.MustNewTables(
		csdb.WithTable(TableIndexStore, "store"),
		csdb.WithTable(TableIndexGroup, "store_group"),
		csdb.WithTable(TableIndexWebsite, "store_website"),
	)
	assert.Equal(t, tm1.Len(), 3)
}

func TestNewTableServicePanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			assert.True(t, errors.IsNotValid(err), "%+v", err)
		} else {
			t.Error("Expecting a panic")
		}
	}()

	_ = csdb.MustNewTables(
		csdb.WithTable(0, ""),
	)
}

func TestTables_Upsert_Insert(t *testing.T) {
	t.Parallel()

	ts := csdb.MustNewTables()

	t.Run("Insert OK", func(t *testing.T) {
		assert.NoError(t, ts.Upsert(0, csdb.NewTable("test1")))
		assert.Equal(t, ts.Len(), 1)
	})
}

func TestTables_DeleteFromCache(t *testing.T) {
	t.Parallel()

	ts := csdb.MustNewTables(csdb.WithTableNames([]int{3, 5, 7}, []string{"a3", "b5", "c7"}))
	t.Run("Delete One", func(t *testing.T) {
		ts.DeleteFromCache(5)
		assert.Exactly(t, 2, ts.Len())
	})
	t.Run("Delete All does nothing", func(t *testing.T) {
		ts.DeleteFromCache()
		assert.Exactly(t, 2, ts.Len())
	})
}

func TestTables_DeleteAllFromCache(t *testing.T) {
	t.Parallel()

	ts := csdb.MustNewTables(csdb.WithTableNames([]int{3, 5, 7}, []string{"a3", "b5", "c7"}))
	ts.DeleteAllFromCache()
	assert.Exactly(t, 0, ts.Len())

}

func TestTables_Upsert_Update(t *testing.T) {
	t.Parallel()

	ts := csdb.MustNewTables(csdb.WithTableNames([]int{3, 5, 7}, []string{"a3", "b5", "c7"}))
	t.Run("One", func(t *testing.T) {
		ts.Upsert(5, csdb.NewTable("x5"))
		assert.Exactly(t, 3, ts.Len())
		tb, err := ts.Table(5)
		assert.NoError(t, err, "%+v", err)
		assert.Exactly(t, `x5`, tb.Name)
	})
}

func TestTables_MustTable(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			assert.True(t, errors.IsNotFound(err), "%+v", err)
		} else {
			t.Error("Expecting a panic")
		}
	}()

	ts := csdb.MustNewTables(csdb.WithTableNames([]int{3}, []string{"a3"}))
	tbl := ts.MustTable(3)
	assert.NotNil(t, tbl)
	tbl = ts.MustTable(44)
	assert.Nil(t, tbl)
}

func TestWithTableNames(t *testing.T) {
	t.Parallel()

	ts := csdb.MustNewTables(csdb.WithTableNames([]int{3, 5, 7}, []string{"a3", "b5", "c7"}))
	t.Run("Ok", func(t *testing.T) {
		assert.Exactly(t, "a3", ts.Name(3))
		assert.Exactly(t, "b5", ts.Name(5))
		assert.Exactly(t, "c7", ts.Name(7))
	})
	t.Run("Imbalanced Length", func(t *testing.T) {
		err := ts.Options(csdb.WithTableNames(nil, []string{"x1"}))
		assert.True(t, errors.IsNotValid(err), "%+v", err)
	})
	t.Run("Invalid Identifier", func(t *testing.T) {
		err := ts.Options(csdb.WithTableNames([]int{1}, []string{"x1"}))
		assert.True(t, errors.IsNotValid(err), "%+v", err)
		assert.Contains(t, err.Error(), `Invalid character "\uf8ff" in name "x\uf8ff1"`)
	})
}

func TestWithLoadTableNames(t *testing.T) {
	t.Parallel()

	dbc, dbMock := cstesting.MockDB(t)
	defer func() {
		dbMock.ExpectClose()
		assert.NoError(t, dbc.Close())
		if err := dbMock.ExpectationsWereMet(); err != nil {
			t.Error("there were unfulfilled expections", err)
		}
	}()

	rows := sqlmock.NewRows([]string{"Tables_in_magento2"}).FromCSVString(`admin_passwords
admin_system_messages
admin_user
admin_user_session
adminnotification_inbox
authorization_role
authorization_rule`)
	dbMock.ExpectQuery("SHOW TABLES LIKE 'catalog_product%'").WillReturnRows(rows)

	tbls, err := csdb.NewTables(csdb.WithLoadTableNames(dbc.DB, "catalog_'product%"))
	assert.NoError(t, err, "%+v", err)

	want := []string{"admin_passwords", "admin_system_messages", "admin_user", "admin_user_session", "adminnotification_inbox", "authorization_role", "authorization_rule"}
	sort.Strings(want)
	have := tbls.Tables()
	sort.Strings(have)

	assert.Exactly(t, want, have)
}

func TestWithLoadColumnDefinitions_Integration(t *testing.T) {
	t.Parallel()

	dbc, mv := cstesting.MustConnectDB()
	if dbc == nil {
		t.Skip("Environment DB DSN not found")
	}
	defer dbc.Close()
	if mv != magento.Version2 {
		t.Skip("Requirement")
	}

	i := 4711
	tm0 := csdb.MustNewTables(
		csdb.WithLoadColumnDefinitions(context.TODO(), dbc.DB),
		csdb.WithTable(i, "admin_user"),
	)

	table, err2 := tm0.Table(i)
	assert.NoError(t, err2)
	assert.Equal(t, 1, table.CountPK)
	assert.Equal(t, 1, table.CountUnique)
	assert.True(t, len(table.Columns.FieldNames()) >= 15)
	// t.Logf("%+v", table.Columns)
}

func TestWithLoadColumnDefinitions2(t *testing.T) {
	t.Parallel()

	dbc, dbMock := cstesting.MockDB(t)
	defer func() {
		dbMock.ExpectClose()
		assert.NoError(t, dbc.Close())
		if err := dbMock.ExpectationsWereMet(); err != nil {
			t.Error("there were unfulfilled expections", err)
		}
	}()

	rows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "ORDINAL_POSITION", "COLUMN_DEFAULT", "IS_NULLABLE", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE", "COLUMN_TYPE", "COLUMN_KEY", "EXTRA", "COLUMN_COMMENT"}).
		FromCSVString(
			`"admin_user","user_id",1,0,"NO","int",0,10,0,"int(10) unsigned","PRI","auto_increment","User ID"
"admin_user","firsname",2,NULL,"YES","varchar",32,0,0,"varchar(32)","","","User First Name"
"admin_user","modified",8,"CURRENT_TIMESTAMP","NO","timestamp",0,0,0,"timestamp","","on update CURRENT_TIMESTAMP","User Modified Time"
`)

	dbMock.ExpectQuery("SELECT.+FROM information_schema.COLUMNS WHERE.+TABLE_NAME IN.+").
		WithArgs("admin_user").
		WillReturnRows(rows)

	i := 4711
	tm0 := csdb.MustNewTables(
		csdb.WithTable(i, "admin_user"),
		csdb.WithLoadColumnDefinitions(context.TODO(), dbc.DB),
	)

	table, err2 := tm0.Table(i)
	assert.NoError(t, err2)
	assert.Equal(t, 1, table.CountPK)
	assert.Equal(t, 0, table.CountUnique)
	assert.Exactly(t, []string{"user_id", "firsname", "modified"}, table.Columns.FieldNames())
	//t.Log(table.Columns.GoString())
}

func TestMustInitTables(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(*testing.T) {
		var ts *csdb.Tables
		ts = csdb.MustInitTables(ts, csdb.WithTableNames([]int{3, 5, 7}, []string{"a3", "b5", "c7"}))
		require.NotNil(t, ts)
		assert.Exactly(t, "a3", ts.Name(3))
		assert.Exactly(t, "b5", ts.Name(5))
		assert.Exactly(t, "c7", ts.Name(7))
	})
	t.Run("1. panic nil new", func(*testing.T) {
		defer func() {
			if r := recover(); r != nil {
				err := r.(error)
				assert.True(t, errors.IsNotValid(err), "%+v", err)
			} else {
				t.Error("Expecting a panic")
			}
		}()
		var ts *csdb.Tables
		csdb.MustInitTables(ts, csdb.WithTableNames([]int{1, 2, 3}, []string{"a3"}))
	})
	t.Run("2. panic nil options", func(*testing.T) {
		defer func() {
			if r := recover(); r != nil {
				err := r.(error)
				assert.True(t, errors.IsNotValid(err), "%+v", err)
			} else {
				t.Error("Expecting a panic")
			}
		}()
		ts := csdb.MustNewTables()
		csdb.MustInitTables(ts, csdb.WithTableNames([]int{1, 2, 3}, []string{"a3"}))
	})
}

func TestWithTableDMLListeners(t *testing.T) {
	t.Parallel()

	counter := 0
	ev := dbr.MustNewListenerBucket(
		dbr.Listen{
			Name:       "l1",
			EventType:  dbr.OnBeforeToSQL,
			SelectFunc: func(_ *dbr.Select) { counter++ },
		},
		dbr.Listen{
			Name:       "l2",
			EventType:  dbr.OnBeforeToSQL,
			SelectFunc: func(_ *dbr.Select) { counter++ },
		},
	)

	t.Run("Nil Table / No-WithTable", func(*testing.T) {
		ts := csdb.MustNewTables(
			csdb.WithTableDMLListeners(33, ev, ev),
			csdb.WithTable(33, "tableA"),
		) // +=2
		tbl := ts.MustTable(33)
		sel := dbr.NewSelect().From("tableA")
		sel.Listeners.Merge(tbl.Listeners.Select) // +=2

		sel.Columns = []string{"a", "b"}
		assert.Exactly(t, "SELECT a, b FROM `tableA`", sel.String())
		assert.Exactly(t, 4, counter) // yes 4 is correct
	})

	t.Run("Non Nil Table", func(*testing.T) {
		ts := csdb.MustNewTables(
			csdb.WithTable(33, "TeschtT", &csdb.Column{Field: "col1"}),
			csdb.WithTableDMLListeners(33, ev, ev),
		) // +=2
		tbl := ts.MustTable(33)
		require.Exactly(t, "TeschtT", tbl.Name)

		sel := tbl.Select()
		require.NotNil(t, sel)
		sel.Listeners.Merge(tbl.Listeners.Select) // +=2

		assert.Exactly(t, "SELECT `main_table`.`col1` FROM `TeschtT` AS `main_table`", sel.String())
		assert.Exactly(t, 8, counter)
	})

	t.Run("Nil Table and after WithTable call", func(*testing.T) {
		ts := csdb.MustNewTables(
			csdb.WithTableDMLListeners(33, ev, ev),
			csdb.WithTable(33, "TeschtU", &csdb.Column{Field: "col1"}),
		) // +=2
		tbl := ts.MustTable(33)
		require.Exactly(t, "TeschtU", tbl.Name)

		sel := tbl.Select()
		require.NotNil(t, sel)
		sel.Listeners.Merge(tbl.Listeners.Select) // +=2

		assert.Exactly(t, "SELECT `main_table`.`col1` FROM `TeschtU` AS `main_table`", sel.String())
		assert.Exactly(t, 12, counter)
	})
}

func TestWithTableLoadColumns(t *testing.T) {
	t.Parallel()

	t.Run("Invalid Identifier", func(t *testing.T) {
		tbls, err := csdb.NewTables(csdb.WithTableLoadColumns(context.TODO(), nil, 0, "H€llo"))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsNotValid(err), "%+v", err)
	})

	t.Run("Ok", func(t *testing.T) {

		dbc, dbMock := cstesting.MockDB(t)
		defer func() {
			dbMock.ExpectClose()
			assert.NoError(t, dbc.Close())
			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Error("there were unfulfilled expections", err)
			}
		}()

		rows := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME", "ORDINAL_POSITION", "COLUMN_DEFAULT", "IS_NULLABLE", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE", "COLUMN_TYPE", "COLUMN_KEY", "EXTRA", "COLUMN_COMMENT"}).
			FromCSVString(
				`"admin_user","user_id",1,0,"NO","int",0,10,0,"int(10) unsigned","PRI","auto_increment","User ID"
	"admin_user","firsname",2,NULL,"YES","varchar",32,0,0,"varchar(32)","","","User First Name"
	"admin_user","modified",8,"CURRENT_TIMESTAMP","NO","timestamp",0,0,0,"timestamp","","on update CURRENT_TIMESTAMP","User Modified Time"
	`)

		dbMock.ExpectQuery("SELECT.+FROM information_schema.COLUMNS WHERE.+TABLE_NAME IN.+").
			WithArgs("admin_user").
			WillReturnRows(rows)

		i := 4711
		tm0 := csdb.MustNewTables(
			csdb.WithTableLoadColumns(context.TODO(), dbc.DB, i, "admin_user"),
		)

		table, err2 := tm0.Table(i)
		assert.NoError(t, err2)
		assert.Equal(t, 1, table.CountPK)
		assert.Equal(t, 0, table.CountUnique)
	})
}

func TestWithObjectFromQuery(t *testing.T) {

	t.Run("Invalid type", func(t *testing.T) {
		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), nil, "proc", 0, "asdasd", "SELECT * from"))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsUnavailable(err), "%+v", err)
	})

	t.Run("Invalid object name", func(t *testing.T) {
		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), nil, "proc", 0, "asdasd", "SELECT * from"))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsNotValid(err), "%+v", err)
	})

	t.Run("drop table fails", func(t *testing.T) {
		dbc, dbMock := cstesting.MockDB(t)
		defer func() {
			dbMock.ExpectClose()
			assert.NoError(t, dbc.Close())
			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Error("there were unfulfilled expections", err)
			}
		}()

		xErr := errors.NewAlreadyClosedf("Connection already closed")
		dbMock.ExpectExec("DROP TABLE IF EXISTS `testTable`").WillReturnError(xErr)

		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), dbc.DB, "table", 0, "testTable", "SELECT * FROM catalog_product_entity", true))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsAlreadyClosed(err), "%+v", err)
	})

	t.Run("create table fails", func(t *testing.T) {
		dbc, dbMock := cstesting.MockDB(t)
		defer func() {
			dbMock.ExpectClose()
			assert.NoError(t, dbc.Close())
			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Error("there were unfulfilled expections", err)
			}
		}()

		xErr := errors.NewAlreadyClosedf("Connection already closed")
		dbMock.ExpectExec(cstesting.SQLMockQuoteMeta("CREATE TABLE `testTable` AS SELECT * FROM catalog_product_entity")).WillReturnError(xErr)

		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), dbc.DB, "table", 0, "testTable", "SELECT * FROM catalog_product_entity", false))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsAlreadyClosed(err), "%+v", err)
	})

	t.Run("load columns fails", func(t *testing.T) {
		dbc, dbMock := cstesting.MockDB(t)
		defer func() {
			dbMock.ExpectClose()
			assert.NoError(t, dbc.Close())
			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Error("there were unfulfilled expections", err)
			}
		}()

		xErr := errors.NewAlreadyClosedf("Connection already closed")
		dbMock.
			ExpectExec(cstesting.SQLMockQuoteMeta("CREATE TABLE `testTable` AS SELECT * FROM catalog_product_entity")).
			WillReturnResult(sqlmock.NewResult(0, 0))

		dbMock.ExpectQuery("SELEC.+\\s+FROM\\s+information_schema\\.COLUMNS").WillReturnError(xErr)

		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), dbc.DB, "table", 0, "testTable", "SELECT * FROM catalog_product_entity", false))
		assert.Nil(t, tbls)
		assert.True(t, errors.IsAlreadyClosed(err), "%+v", err)
	})

	t.Run("create view", func(t *testing.T) {
		dbc, dbMock := cstesting.MockDB(t)
		defer func() {
			dbMock.ExpectClose()
			assert.NoError(t, dbc.Close())
			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Error("there were unfulfilled expections", err)
			}
		}()

		dbMock.
			ExpectExec(cstesting.SQLMockQuoteMeta("CREATE VIEW `testTable` AS SELECT * FROM core_config_data")).
			WillReturnResult(sqlmock.NewResult(0, 0))

		dbMock.ExpectQuery("SELECT.+FROM information_schema.COLUMNS WHERE").
			WithArgs("testTable").
			WillReturnRows(
				cstesting.MustMockRows(cstesting.WithFile("testdata/core_config_data_columns.csv")))

		tbls, err := csdb.NewTables(csdb.WithTableOrViewFromQuery(context.TODO(), dbc.DB, "view", 10, "testTable", "SELECT * FROM core_config_data", false))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		assert.Exactly(t, "testTable", tbls.MustTable(10).Name)
		assert.True(t, tbls.MustTable(10).IsView, "Table should be a view")
	})

}
