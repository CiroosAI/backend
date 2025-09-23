package utils

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

// Minimal internal validator to avoid external dependency. Supports:
// - required
// - phone8 (starts with '8' and 8-13 digits total)
// - nameok (letters, numbers, space, hyphen, apostrophe, 1-100 chars)
// - pwdmin (min length 6)
// - eqfield=OtherField (field equals another field)

var (
	rePhone8 = regexp.MustCompile(`^8[0-9]{7,12}$`)
	reNameOK = regexp.MustCompile(`^[A-Za-z0-9 \-']{1,100}$`)
)

// ValidateStruct inspects struct tags `validate:"..."` and returns the first error encountered.
func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("ValidateStruct expects a struct or pointer to struct")
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}
		parts := strings.Split(tag, ",")
		fv := v.Field(i)
		var sval string
		if fv.IsValid() && fv.Kind() == reflect.String {
			sval = fv.String()
		}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "required" {
				if sval == "" {
					return errors.New(field.Name + " is required")
				}
			} else if p == "phone8" {
				if sval != "" && !rePhone8.MatchString(sval) {
					return errors.New(field.Name + " must be an Indonesian local phone number starting with 8")
				}
			} else if p == "nameok" {
				if sval != "" && !reNameOK.MatchString(sval) {
					return errors.New(field.Name + " contains invalid characters")
				}
			} else if p == "pwdmin" {
				if len(sval) < 6 {
					return errors.New(field.Name + " must be at least 6 characters")
				}
			} else if strings.HasPrefix(p, "eqfield=") {
				other := strings.TrimPrefix(p, "eqfield=")
				of := v.FieldByName(other)
				if of.IsValid() && of.Kind() == reflect.String {
					if sval != of.String() {
						return errors.New(field.Name + " must equal " + other)
					}
				}
			}
		}
	}
	return nil
}
