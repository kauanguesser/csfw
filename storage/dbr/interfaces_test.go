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

package dbr_test

import (
	"context"
	"database/sql"

	"github.com/corestoreio/csfw/storage/dbr"
)

var _ dbr.DBer = (*sql.DB)(nil)
var _ dbr.Preparer = (*sql.DB)(nil)
var _ dbr.Querier = (*sql.DB)(nil)
var _ dbr.Execer = (*sql.DB)(nil)
var _ dbr.QueryRower = (*sql.DB)(nil)

var _ dbr.Stmter = (*sql.Stmt)(nil)
var _ dbr.StmtQueryRower = (*sql.Stmt)(nil)
var _ dbr.StmtQueryer = (*sql.Stmt)(nil)
var _ dbr.StmtExecer = (*sql.Stmt)(nil)

var _ dbr.Txer = (*sql.Tx)(nil)

var _ dbr.Preparer = (*dbMock)(nil)
var _ dbr.Querier = (*dbMock)(nil)
var _ dbr.Execer = (*dbMock)(nil)

type dbMock struct {
	error
	prepareFn func(query string) (*sql.Stmt, error)
}

func (pm dbMock) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if pm.error != nil {
		return nil, pm.error
	}
	return pm.prepareFn(query)
}

func (pm dbMock) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if pm.error != nil {
		return nil, pm.error
	}
	return nil, nil
}

func (pm dbMock) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if pm.error != nil {
		return nil, pm.error
	}
	return nil, nil
}
