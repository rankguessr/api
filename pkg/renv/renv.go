package renv

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	ErrInvalidType = errors.New("invalid type: expected pointer to struct")
	ErrInvalidTag  = errors.New("invalid tag: expected format 'env:VAR_NAME[,options]'")
)

func Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrInvalidType
	}

	for i := 0; i < rv.Elem().NumField(); i++ {
		field := rv.Elem().Type().Field(i)

		// skip nested structs for now
		if field.Type.Kind() == reflect.Struct {
			continue
		}

		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			return ErrInvalidTag
		}

		varName := parts[0]
		required := false
		if len(parts) > 1 {
			for _, option := range parts[1:] {
				if option == "required" {
					required = true
				}
			}
		}

		value, exists := os.LookupEnv(varName)
		if !exists {
			if required {
				return fmt.Errorf("%s environment variable is required but not set", varName)
			}

			continue
		}

		// TODO: add support for other types
		elemField := rv.Elem().Field(i)
		if elemField.CanSet() {
			elemField.SetString(value)
		}
	}

	return nil
}
