// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"regexp"
	"strings"
)

var validId = regexp.MustCompile("^[a-z]([a-z0-9-]?[a-z0-9])*$")
var splCharMatch = regexp.MustCompile("[^a-zA-Z0-9-]")
var firstCapMatch = regexp.MustCompile("(.)([A-Z][a-z]+)")
var allCapMatch = regexp.MustCompile("([a-z0-9])([A-Z])")

func IsValidIdentifier(identifier string) bool {
	match := validId.MatchString(identifier)
	return match
}

func ConvertIdentifier(identifier string) string {
	var flag bool
	//Removing Special characters from the identifier
	id := splCharMatch.ReplaceAllString(identifier, "")

	//Replacing all capital characters with `-`+<Cap char>
	id = firstCapMatch.ReplaceAllString(id, "${1}-${2}")
	id = allCapMatch.ReplaceAllString(id, "${1}-${2}")

	//Removing `digit` and `-` from prefix
	for i := 0; i < len(id); i++ {
		asciiValue := int(id[i])
		if (asciiValue >= 48 && asciiValue <= 57) || asciiValue == 45 {
			flag = true
			continue
		} else {
			if flag {
				id = id[i:]
			}
			break
		}
	}

	return strings.ToLower(id)
}
