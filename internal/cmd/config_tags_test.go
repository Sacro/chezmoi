package cmd

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var expectedTags = []string{"json", "yaml", "mapstructure"}

func TestExportedFieldsHaveMatchingMarshalTags(t *testing.T) {
	failed, errmsg := verifyTagsArePresentAndMatch(reflect.TypeFor[ConfigFile]())
	if failed {
		t.Error(errmsg)
	}
}

func fieldTypesNeedsVerification(ft reflect.Type) []reflect.Type {
	kind := ft.Kind()
	if kind < reflect.Array || kind == reflect.String { // its a ~scalar type
		return []reflect.Type{}
	} else if kind == reflect.Struct {
		return []reflect.Type{ft}
	}
	switch kind {
	case reflect.Pointer:
		fallthrough
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return fieldTypesNeedsVerification(ft.Elem())
	case reflect.Map:
		return append(fieldTypesNeedsVerification(ft.Key()), fieldTypesNeedsVerification(ft.Elem())...)
	default:
		return []reflect.Type{} // ... we'll assume interface types, funcs, chans are okay.
	}
}

func verifyTagsArePresentAndMatch(structType reflect.Type) (failed bool, errmsg string) {
	name := structType.Name()
	fields := reflect.VisibleFields(structType)
	failed = false

	var errs strings.Builder

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		ts := f.Tag
		tagValueGroups := make(map[string][]string)

		for _, tagName := range expectedTags {
			tagValue, tagPresent := ts.Lookup(tagName)

			if !tagPresent {
				errs.WriteString(fmt.Sprintf("\n%s field %s is missing a `%s:` tag", name, f.Name, tagName))
				failed = true
			}

			matchingTags, notFirstOccurrence := tagValueGroups[tagValue]
			if notFirstOccurrence {
				tagValueGroups[tagValue] = append(matchingTags, tagName)
			} else {
				tagValueGroups[tagValue] = []string{tagName}
			}
		}

		if len(tagValueGroups) > 1 {
			errs.WriteString(fmt.Sprintf("\n%s field %s has non-matching tag names:", name, f.Name))

			for value, tagsMatching := range tagValueGroups {
				if len(tagsMatching) == 1 {
					errs.WriteString(fmt.Sprintf("\n    %s says \"%s\"", tagsMatching[0], value))
				} else {
					errs.WriteString(fmt.Sprintf("\n    (%s) each say \"%s\"", strings.Join(tagsMatching, ", "), value))
				}
			}
			failed = true
		}

		verifyTypes := fieldTypesNeedsVerification(f.Type)
		for _, ft := range verifyTypes {
			subFailed, suberrs := verifyTagsArePresentAndMatch(ft)
			if subFailed {
				errs.WriteString(fmt.Sprintf("\n In %s.%s:", name, f.Name))
				errs.WriteString(strings.ReplaceAll(suberrs, "\n", "\n    "))
				failed = true
			}
		}
	}

	return failed, errs.String()
}
