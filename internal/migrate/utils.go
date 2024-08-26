// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migrate

import (
	"fmt"
	"strings"
)

const maxIdentifierLength = 100

// DisplayNameToIdentifier converts display name to a unique identifier.
func DisplayNameToIdentifier(displayName, prefix, suffix string) string {
	const placeholder = '_'
	const specialChars = ".-_"
	// remove / replace any illegal characters
	// Identifier Regex: ^[a-zA-Z0-9-_.]*$
	identifier := strings.Map(func(r rune) rune {
		switch {
		// drop any control characters or empty characters
		case r < 32 || r == 127:
			return -1

		// keep all allowed character
		case ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') ||
			strings.ContainsRune(specialChars, r):
			return r

		// everything else is replaced with the placeholder
		default:
			return placeholder
		}
	}, displayName)

	// remove any leading/trailing special characters
	identifier = strings.Trim(identifier, specialChars)

	// ensure string doesn't start with numbers (leading '_' is valid)
	if len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9' {
		identifier = string(placeholder) + identifier
	}

	// remove consecutive special characters
	identifier = sanitizeConsecutiveChars(identifier, specialChars)

	// ensure length restrictions
	if len(identifier) > maxIdentifierLength {
		identifier = identifier[0:maxIdentifierLength]
	}

	if len(identifier) == 0 {
		return fmt.Sprintf("%s%c%s", prefix, placeholder, suffix)
	}

	// adding suffix to make sure the identifier would be unique
	return fmt.Sprintf("%s%c%s", identifier, placeholder, suffix)
}

func sanitizeConsecutiveChars(in string, charSet string) string {
	if len(in) == 0 {
		return ""
	}

	inSet := func(b byte) bool {
		return strings.ContainsRune(charSet, rune(b))
	}

	out := strings.Builder{}
	out.WriteByte(in[0])
	for i := 1; i < len(in); i++ {
		if inSet(in[i]) && inSet(in[i-1]) {
			continue
		}
		out.WriteByte(in[i])
	}

	return out.String()
}
