package dbr

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/corestoreio/errors"
	"github.com/stretchr/testify/assert"
)

var (
	stringJSON      = []byte(`"test"`)
	blankStringJSON = []byte(`""`)
	nullStringJSON  = []byte(`{"NullString":"test","Valid":true}`)

	nullJSON    = []byte(`null`)
	invalidJSON = []byte(`:)`)
)

type stringInStruct struct {
	Test NullString `json:"test,omitempty"`
}

func TestStringFrom(t *testing.T) {
	t.Parallel()
	str := MakeNullString("test")
	assertStr(t, str, "MakeNullString() string")

	zero := MakeNullString("")
	if !zero.Valid {
		t.Error("MakeNullString(0)", "is invalid, but should be valid")
	}
}

func TestUnmarshalString(t *testing.T) {
	t.Parallel()
	var str NullString
	maybePanic(json.Unmarshal(stringJSON, &str))
	assertStr(t, str, "string json")

	var ns NullString
	maybePanic(json.Unmarshal(nullStringJSON, &ns))
	assertStr(t, ns, "sql.NullString json")

	var blank NullString
	maybePanic(json.Unmarshal(blankStringJSON, &blank))
	if !blank.Valid {
		t.Error("blank string should be valid")
	}

	var null NullString
	maybePanic(json.Unmarshal(nullJSON, &null))
	assertNullStr(t, null, "null json")

	var badType NullString
	err := json.Unmarshal(boolJSON, &badType)
	if err == nil {
		panic("err should not be nil")
	}
	assertNullStr(t, badType, "wrong type json")

	var invalid NullString
	err = invalid.UnmarshalJSON(invalidJSON)
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("expected json.SyntaxError, not %T", err)
	}
	assertNullStr(t, invalid, "invalid json")
}

func TestTextUnmarshalString(t *testing.T) {
	t.Parallel()
	var str NullString
	err := str.UnmarshalText([]byte("test"))
	maybePanic(err)
	assertStr(t, str, "UnmarshalText() string")

	var null NullString
	err = null.UnmarshalText([]byte(""))
	maybePanic(err)
	assertNullStr(t, null, "UnmarshalText() empty string")

	var iv NullString
	err = iv.UnmarshalText([]byte{0x44, 0xff, 0x01})
	assert.True(t, errors.IsNotValid(err), "%+v", err)
}

func TestMarshalString(t *testing.T) {
	t.Parallel()
	str := MakeNullString("test")
	data, err := json.Marshal(str)
	maybePanic(err)
	assertJSONEquals(t, data, `"test"`, "non-empty json marshal")
	data, err = str.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "test", "non-empty text marshal")

	// empty values should be encoded as an empty string
	zero := MakeNullString("")
	data, err = json.Marshal(zero)
	maybePanic(err)
	assertJSONEquals(t, data, `""`, "empty json marshal")
	data, err = zero.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "", "string marshal text")

	zero.Valid = false
	data, err = zero.MarshalText()
	maybePanic(err)
	assert.Exactly(t, []byte{}, data)
}

func TestStringPointer(t *testing.T) {
	t.Parallel()
	str := MakeNullString("test")
	ptr := str.Ptr()
	if *ptr != "test" {
		t.Errorf("bad %s string: %#v ≠ %s\n", "pointer", ptr, "test")
	}

	null := MakeNullString("", false)
	ptr = null.Ptr()
	if ptr != nil {
		t.Errorf("bad %s string: %#v ≠ %s\n", "nil pointer", ptr, "nil")
	}
}

func TestStringIsZero(t *testing.T) {
	t.Parallel()
	str := MakeNullString("test")
	if str.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	blank := MakeNullString("")
	if blank.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	empty := MakeNullString("", true)
	if empty.IsZero() {
		t.Errorf("IsZero() should be false")
	}
}

func TestStringSetValid(t *testing.T) {
	t.Parallel()
	change := MakeNullString("", false)
	assertNullStr(t, change, "SetValid()")
	change.SetValid("test")
	assertStr(t, change, "SetValid()")
}

func TestStringScan(t *testing.T) {
	t.Parallel()
	var str NullString
	err := str.Scan("test")
	maybePanic(err)
	assertStr(t, str, "scanned string")

	var null NullString
	err = null.Scan(nil)
	maybePanic(err)
	assertNullStr(t, null, "scanned null")
}

func maybePanic(err error) {
	if err != nil {
		panic(err)
	}
}

var _ fmt.GoStringer = (*NullString)(nil)

func TestString_GoString(t *testing.T) {
	t.Parallel()
	s := MakeNullString("test", true)
	assert.Exactly(t, "dbr.MakeNullString(`test`)", s.GoString())

	s = MakeNullString("test", false)
	assert.Exactly(t, "dbr.NullString{}", s.GoString())

	s = MakeNullString("te`st", true)
	gsWant := []byte("dbr.MakeNullString(`te`+\"`\"+`st`)")
	if !bytes.Equal(gsWant, []byte(s.GoString())) {
		t.Errorf("Have: %#v Want: %v", s.GoString(), string(gsWant))
	}
}

func assertStr(t *testing.T, s NullString, from string) {
	if s.String != "test" {
		t.Errorf("bad %s string: %s ≠ %s\n", from, s.String, "test")
	}
	if !s.Valid {
		t.Error(from, "is invalid, but should be valid")
	}
}

func assertNullStr(t *testing.T, s NullString, from string) {
	if s.Valid {
		t.Error(from, "is valid, but should be invalid")
	}
}

func assertJSONEquals(t *testing.T, data []byte, cmp string, from string) {
	if string(data) != cmp {
		t.Errorf("bad %s data: %s ≠ %s\n", from, data, cmp)
	}
}

func TestNullString_Argument(t *testing.T) {
	t.Parallel()

	nss := []NullString{
		{
			NullString: sql.NullString{
				String: "Gopher",
			},
		},
		{
			NullString: sql.NullString{
				String: "Ru'st\")y('",
				Valid:  true,
			},
		},
	}
	var buf bytes.Buffer
	args := make([]interface{}, 0, 2)
	for i, ns := range nss {
		ns.toIFace(&args)
		ns.writeTo(&buf, i)

		arg := ns.Operator(NotBetween)
		assert.Exactly(t, NotBetween, arg.operator(), "Index %d", i)
		assert.Exactly(t, 1, arg.len(), "Length must be always one")
	}
	assert.Exactly(t, []interface{}{interface{}(nil), "Ru'st\")y('"}, args)
	assert.Exactly(t, "NULL'Ru\\'st\\\")y(\\''", buf.String())
}

func TestArgNullString(t *testing.T) {
	t.Parallel()

	args := ArgNullString(MakeNullString("1'; DROP TABLE users-- 1"), MakeNullString("Rusty", false), MakeNullString("Powerلُلُصّبُلُلصّبُررً ॣ ॣh ॣ ॣ冗"))
	assert.Exactly(t, 3, args.len())
	args = args.Operator(NotIn)
	assert.Exactly(t, 1, args.len())

	t.Run("IN operator", func(t *testing.T) {
		args = args.Operator(In)
		var buf bytes.Buffer
		argIF := make([]interface{}, 0, 2)
		if err := args.writeTo(&buf, 0); err != nil {
			t.Fatalf("%+v", err)
		}
		args.toIFace(&argIF)
		assert.Exactly(t, []interface{}{"1'; DROP TABLE users-- 1", interface{}(nil), "Powerلُلُصّبُلُلصّبُررً ॣ ॣh ॣ ॣ冗"}, argIF)
		assert.Exactly(t, "('1\\'; DROP TABLE users-- 1',NULL,'Powerلُلُصّبُلُلصّبُررً ॣ ॣh ॣ ॣ冗')", buf.String())
	})

	t.Run("Not Equal operator", func(t *testing.T) {
		args = args.Operator(NotEqual)
		var buf bytes.Buffer
		argIF := make([]interface{}, 0, 2)
		for i := 0; i < args.len(); i++ {
			if err := args.writeTo(&buf, i); err != nil {
				t.Fatalf("%+v", err)
			}
		}
		args.toIFace(&argIF)
		assert.Exactly(t, []interface{}{"1'; DROP TABLE users-- 1", interface{}(nil), "Powerلُلُصّبُلُلصّبُررً ॣ ॣh ॣ ॣ冗"}, argIF)
		assert.Exactly(t, "'1\\'; DROP TABLE users-- 1'NULL'Powerلُلُصّبُلُلصّبُررً ॣ ॣh ॣ ॣ冗'", buf.String())
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		var buf bytes.Buffer

		args := ArgNullString(MakeNullString("\x00\xff"))
		err := args.writeTo(&buf, 0)
		assert.True(t, errors.IsNotValid(err), "%+v", err)
		buf.Reset()

		args = ArgNullString(MakeNullString("\x00\xff"), MakeNullString("2nd"))
		err = args.writeTo(&buf, 0)
		assert.True(t, errors.IsNotValid(err), "%+v", err)
		buf.Reset()

		args = args.Operator(In)
		err = args.writeTo(&buf, -1)
		assert.True(t, errors.IsNotValid(err), "%+v", err)
	})

	t.Run("single arg", func(t *testing.T) {
		args = ArgNullString(MakeNullString("1';"))
		args = args.Operator(NotEqual)
		var buf bytes.Buffer
		argIF := make([]interface{}, 0, 2)
		for i := 0; i < args.len(); i++ {
			if err := args.writeTo(&buf, i); err != nil {
				t.Fatalf("%+v", err)
			}
		}
		args.toIFace(&argIF)
		assert.Exactly(t, []interface{}{"1';"}, argIF)
		assert.Exactly(t, "'1\\';'", buf.String())
	})

}
