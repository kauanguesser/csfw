// Copyright 2015-2017, Cyrill @ Schumacher.fm and the CoreStore contributors
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
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/errors"
)

var _ dbr.ArgumentGenerater = (*dbrPerson)(nil)

type dbrPerson struct {
	ID    int64 `db:"id"`
	Name  string
	Email dbr.NullString
	Key   dbr.NullString
}

func (p *dbrPerson) GenerateArguments(statementType byte, columns, condition []string) (dbr.Arguments, error) {
	if statementType != dbr.StatementTypeUpdate {
		return nil, errors.NewNotSupportedf("[dbr] Statement Type %q not yet supported", statementType)
	}

	args := make(dbr.Arguments, 0, 4) // 4 == number of fields in the struct
	for _, c := range columns {
		switch c {
		case "name":
			args = append(args, dbr.ArgString(p.Name))
		case "email":
			args = append(args, dbr.ArgNullString(p.Email))
		case "key":
			args = append(args, dbr.ArgNullString(p.Key))
		default:
			return nil, errors.NewNotFoundf("[dbr_test] Column %q not found", c)
		}
	}
	for _, c := range condition {
		switch c {
		case "id":
			args = append(args, dbr.ArgInt64(p.ID))
		}
	}
	return args, nil
}
