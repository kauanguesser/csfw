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

package dbr

import (
	"context"
	"database/sql"
)

// DBer is a composition of multiple interfaces to describe the common needed
// behaviour for querying a database. This interface is context independent.
type DBer interface {
	Preparer
	Execer
	Querier
	QueryRower
}

// Preparer creates a new prepared statement.
type Preparer interface {
	// PrepareContext creates a prepared statement for later queries or
	// executions. Multiple queries or executions may be run concurrently from
	// the returned statement. The caller must call the statement's Close method
	// when the statement is no longer needed.
	//
	// The provided context is used for the preparation of the statement, not
	// for the execution of the statement.
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Querier can execute a SELECT query which can return many rows.
type Querier interface {
	// QueryContext executes a query that returns rows, typically a SELECT. The
	// args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Execer can execute all other queries except SELECT.
type Execer interface {
	// ExecContext executes a query that doesn't return rows.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// QueryRower executes a SELECT query which returns one row.
type QueryRower interface {
	// QueryRowContext executes a query that is expected to return at most one
	// row. QueryRow always returns a non-nil value. Errors are deferred until
	// Row'ab Scan method is called.
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Stmter is a composition of multiple interfaces to describe the common needed
// behaviour for querying a database within a prepared statement. This interface
// is context independent.
type Stmter interface {
	StmtExecer
	StmtQueryer
	StmtQueryRower
}

// StmtExecer executes a prepared non-SELECT statement
type StmtExecer interface {
	// ExecContext executes a query that doesn't return rows.
	// For example: an INSERT and UPDATE.
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
}

// StmtQueryer executes a prepared e.g. SELECT statement which can return many
// rows.
type StmtQueryer interface {
	QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
}

// StmtQueryRower executes a prepared e.g. SELECT statement which can return one
// row.
type StmtQueryRower interface {
	QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
}

// TxBeginner starts a transaction
type TxBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Txer is an in-progress database transaction.
//
// A transaction must end with a call to Commit or Rollback.
//
// After a call to Commit or Rollback, all operations on the transaction fail
// with ErrTxDone.
//
// The statements prepared for a transaction by calling the transaction'ab
// Prepare or Stmt methods are closed by the call to Commit or Rollback.
type Txer interface {
	Commit() error
	Rollback() error
	Stmt(stmt *sql.Stmt) *sql.Stmt
	DBer
}

var _ Txer = (*txMock)(nil)

// txMock does nothing and returns always nil
type txMock struct{}

func (txMock) Commit() error                                                       { return nil }
func (txMock) Rollback() error                                                     { return nil }
func (txMock) Stmt(stmt *sql.Stmt) *sql.Stmt                                       { return nil }
func (txMock) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) { return nil, nil }
func (txMock) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (txMock) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (txMock) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
