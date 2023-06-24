package pg

import (
	"fmt"
	"unicode"
)

// copied from https://github.com/lib/pq/blob/2a217b94f5ccd3de31aec4152a541b9ff64bed05/conn.go
// under Copyright (c) 2011-2013, 'pq' Contributors Portions Copyright (C) 2011 Blake Mizerany

type values map[string]string

// scanner implements a tokenizer for libpq-style option strings.
type scanner struct {
	s []rune
	i int
}

// newscanner returns a new scanner initialized with the option string s.
func newscanner(s string) *scanner {
	return &scanner{[]rune(s), 0}
}

// next returns the next rune.
// It returns 0, false if the end of the text has been reached.
func (s *scanner) next() (rune, bool) {
	if s.i >= len(s.s) {
		return 0, false
	}
	r := s.s[s.i]
	s.i++
	return r, true
}

// skipSpaces returns the next non-whitespace rune.
// It returns 0, false if the end of the text has been reached.
func (s *scanner) skipSpaces() (rune, bool) {
	r, ok := s.next()
	for unicode.IsSpace(r) && ok {
		r, ok = s.next()
	}
	return r, ok
}

// ParseOpts parses the options from name and adds them to the values.
//
// The parsing code is based on conninfo_parse from libpq's fe-connect.c
func ParseOpts(name string) (map[string]string, error) {
	s := newscanner(name)

	o := make(values)
	for {
		var (
			keyRunes, valRunes []rune
			r                  rune
			ok                 bool
		)

		if r, ok = s.skipSpaces(); !ok {
			break
		}

		// Scan the key
		for !unicode.IsSpace(r) && r != '=' {
			keyRunes = append(keyRunes, r)
			if r, ok = s.next(); !ok {
				break
			}
		}

		// Skip any whitespace if we're not at the = yet
		if r != '=' {
			r, ok = s.skipSpaces()
		}

		// The current character should be =
		if r != '=' || !ok {
			return nil, fmt.Errorf(`missing "=" after %q in connection info string"`, string(keyRunes))
		}

		// Skip any whitespace after the =
		if r, ok = s.skipSpaces(); !ok {
			// If we reach the end here, the last value is just an empty string as per libpq.
			o[string(keyRunes)] = ""
			break
		}

		if r != '\'' {
			for !unicode.IsSpace(r) {
				if r == '\\' {
					if r, ok = s.next(); !ok {
						return nil, fmt.Errorf(`missing character after backslash`)
					}
				}
				valRunes = append(valRunes, r)

				if r, ok = s.next(); !ok {
					break
				}
			}
		} else {
		quote:
			for {
				if r, ok = s.next(); !ok {
					return nil, fmt.Errorf(`unterminated quoted string literal in connection string`)
				}
				switch r {
				case '\'':
					break quote
				case '\\':
					r, _ = s.next()
					fallthrough
				default:
					valRunes = append(valRunes, r)
				}
			}
		}

		o[string(keyRunes)] = string(valRunes)
	}

	return map[string]string(o), nil
}
