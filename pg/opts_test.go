package pg

import (
	"reflect"
	"testing"
)

// Copied from https://github.com/lib/pq/blob/2a217b94f5ccd3de31aec4152a541b9ff64bed05/conn_test.go
// under Copyright (c) 2011-2013, 'pq' Contributors Portions Copyright (C) 2011 Blake Mizerany

func TestParseOpts(t *testing.T) {
	tests := []struct {
		in       string
		expected values
		valid    bool
	}{
		{"dbname=hello user=goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user=goodbye  ", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname = hello user=goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user =goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user= goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"host=localhost password='correct horse battery staple'", values{"host": "localhost", "password": "correct horse battery staple"}, true},
		{"dbname=データベース password=パスワード", values{"dbname": "データベース", "password": "パスワード"}, true},
		{"dbname=hello user=''", values{"dbname": "hello", "user": ""}, true},
		{"user='' dbname=hello", values{"dbname": "hello", "user": ""}, true},
		// The last option value is an empty string if there's no non-whitespace after its =
		{"dbname=hello user=   ", values{"dbname": "hello", "user": ""}, true},

		// The parser ignores spaces after = and interprets the next set of non-whitespace characters as the value.
		{"user= password=foo", values{"user": "password=foo"}, true},

		// Backslash escapes next char
		{`user=a\ \'\\b`, values{"user": `a '\b`}, true},
		{`user='a \'b'`, values{"user": `a 'b`}, true},

		// Incomplete escape
		{`user=x\`, values{}, false},

		// No '=' after the key
		{"postgre://marko@internet", values{}, false},
		{"dbname user=goodbye", values{}, false},
		{"user=foo blah", values{}, false},
		{"user=foo blah   ", values{}, false},

		// Unterminated quoted value
		{"dbname=hello user='unterminated", values{}, false},
	}

	for _, test := range tests {
		o := make(values)
		err := ParseOpts(test.in, o)

		switch {
		case err != nil && test.valid:
			t.Errorf("%q got unexpected error: %s", test.in, err)
		case err == nil && test.valid && !reflect.DeepEqual(test.expected, o):
			t.Errorf("%q got: %#v want: %#v", test.in, o, test.expected)
		case err == nil && !test.valid:
			t.Errorf("%q expected an error", test.in)
		}
	}
}
