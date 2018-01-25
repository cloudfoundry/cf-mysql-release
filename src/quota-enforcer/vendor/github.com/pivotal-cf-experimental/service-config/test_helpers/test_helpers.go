package test_helpers

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Validator interface {
	Validate() error
}

func setNestedFieldToEmpty(obj interface{}, nestedFieldNames []string) error {

	s := reflect.ValueOf(obj).Elem()
	if s.Type().Kind() == reflect.Slice {
		if s.Len() == 0 {
			return errors.New("Trying to set nested property on empty slice")
		}
		s = s.Index(0)
	}

	currFieldName := nestedFieldNames[0]
	remainingFieldNames := nestedFieldNames[1:]
	field := s.FieldByName(currFieldName)
	if field.IsValid() == false {
		return errors.New(fmt.Sprintf("Field '%s' is not defined", currFieldName))
	}

	if len(remainingFieldNames) == 0 {
		fieldType := field.Type()
		field.Set(reflect.Zero(fieldType))
		return nil
	}
	return setNestedFieldToEmpty(field.Addr().Interface(), remainingFieldNames)
}

func setFieldToEmpty(obj interface{}, fieldName string) error {
	return setNestedFieldToEmpty(obj, strings.Split(fieldName, "."))
}

func IsRequiredField(obj interface{}, fieldName string) error {
	err := setFieldToEmpty(obj, fieldName)
	if err != nil {
		return err
	}

	v, ok := obj.(Validator)
	if !ok {
		return errors.New("object under test does not implement Validator interface")
	}

	err = v.Validate()
	if err == nil {
		return errors.New(fmt.Sprintf("expected Validate to fail when '%s' is empty", fieldName))
	}

	fieldParts := strings.Split(fieldName, ".")
	for _, fieldPart := range fieldParts {
		if !strings.Contains(err.Error(), fieldPart) {
			return errors.New(fmt.Sprintf("expected Validate error\n%s\nto contain\n'%s'",
				err.Error(), fieldPart))
		}
	}

	return nil
}

func IsOptionalField(obj interface{}, fieldName string) error {
	err := setFieldToEmpty(obj, fieldName)
	if err != nil {
		return err
	}

	v, ok := obj.(Validator)
	if !ok {
		return errors.New("object under test does not implement Validator interface")
	}

	err = v.Validate()
	if err != nil {
		return errors.New(fmt.Sprintf("expected Validate to succeed when '%s' is empty\nValidate error: %s", fieldName, err))
	}

	return nil
}
