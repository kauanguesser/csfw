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

package csdb

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDSN(t *testing.T) {
	tests := []struct {
		env        string
		envContent string
		err        error
		returnErr  bool
	}{
		{
			env:        "TEST_CS_1",
			envContent: "Hello",
			err:        errors.New("World"),
			returnErr:  false,
		},
	}

	for _, test := range tests {
		os.Setenv(test.env, test.envContent)
		s, aErr := getDSN(test.env, test.err)
		assert.Equal(t, test.envContent, s)
		assert.NoError(t, aErr)

		s, aErr = getDSN(test.env+"NOTFOUND", test.err)
		assert.Equal(t, "", s)
		assert.Error(t, aErr)
		assert.Equal(t, test.err, aErr)
	}
}

func TestGetParsedDSN(t *testing.T) {
	currentDSN := os.Getenv(EnvDSN)
	defer func() {
		if currentDSN != "" {
			os.Setenv(EnvDSN, currentDSN)
		}
	}()

	tests := []struct {
		envContent string
		wantErr    error
		wantURL    string
	}{
		{"Invalid://\\DSN", errors.New("Cannot parse DSN into URL"), ""},
		{
			"mysql://root:passwrd@localhost:3306/databaseName?BinlogSlaveId=100&BinlogDumpNonBlock=0",
			nil,
			`mysql://root:passw%EF%A3%BFrd@localhost:3306/databaseName?BinlogSlaveId=100&BinlogDumpNonBlock=0`,
		},
	}

	for i, test := range tests {
		os.Setenv(EnvDSN, test.envContent)

		haveURL, haveErr := GetParsedDSN()
		if test.wantErr != nil {
			assert.Nil(t, haveURL)
			require.Error(t, haveErr, "Index %d", i)
			assert.Contains(t, haveErr.Error(), test.wantErr.Error(), "Index %d => %+v", i, haveErr)
			continue
		}
		assert.Exactly(t, test.wantURL, haveURL.String(), "Index %d", i)
	}
}
