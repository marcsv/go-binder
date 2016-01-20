package binder

import (
	"errors"
	"mime/multipart"
	"reflect"
	"strconv"
)

var (
	errUnknownFieldType = errors.New("unknown field type")
)

func formToStruct(form map[string][]string, formFiles map[string][]*multipart.FileHeader, target interface{}) error {
	structType := reflect.TypeOf(target).Elem()
	structValue := reflect.ValueOf(target).Elem()

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if fieldValue.Kind() == reflect.Struct {
			err := formToStruct(form, formFiles, fieldValue.Addr().Interface())
			if err != nil {
				return err
			}

			continue
		}

		fieldKey := fieldType.Tag.Get("form")
		if fieldKey == "" {
			continue
		}

		formValue, exists := form[fieldKey]
		if exists {
			if fieldValue.Kind() == reflect.Slice && len(formValue) > 0 {
				err := setSlice(fieldValue, formValue)
				if err != nil {
					return err
				}
			} else {
				err := setField(fieldType.Type.Kind(), fieldValue, formValue[0])
				if err != nil {
					return err
				}
			}
		}

		formFile, exists := formFiles[fieldKey]
		if exists {
			fileHeaderType := reflect.TypeOf(&multipart.FileHeader{})
			lenFile := len(formFile)

			if fieldValue.Kind() == reflect.Slice && lenFile > 0 && fieldValue.Type().Elem() == fileHeaderType {
				slice := reflect.MakeSlice(fieldValue.Type(), lenFile, lenFile)

				for i := 0; i < lenFile; i++ {
					slice.Index(i).Set(reflect.ValueOf(formFile[i]))
				}

				fieldValue.Set(slice)
			} else if fieldValue.Type() == fileHeaderType {
				fieldValue.Set(reflect.ValueOf(formFile[0]))
			}
		}
	}

	return nil
}

func setField(fieldKind reflect.Kind, field reflect.Value, value string) error {
	switch fieldKind {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		return setBool(field, value)
	case reflect.Int:
		return setInt(field, value, 0)
	case reflect.Int8:
		return setInt(field, value, 8)
	case reflect.Int16:
		return setInt(field, value, 16)
	case reflect.Int32:
		return setInt(field, value, 32)
	case reflect.Int64:
		return setInt(field, value, 64)
	case reflect.Uint:
		return setUint(field, value, 0)
	case reflect.Uint8:
		return setUint(field, value, 8)
	case reflect.Uint16:
		return setUint(field, value, 16)
	case reflect.Uint32:
		return setUint(field, value, 32)
	case reflect.Uint64:
		return setUint(field, value, 64)
	case reflect.Float32:
		return setFloat(field, value, 32)
	case reflect.Float64:
		return setFloat(field, value, 64)
	default:
		return errUnknownFieldType
	}

	return nil
}

func setBool(field reflect.Value, value string) error {
	if value == "" {
		field.SetBool(false)
		return nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolValue)
	}

	return err
}

func setInt(field reflect.Value, value string, size int) error {
	if value == "" {
		field.SetInt(0)
		return nil
	}

	intValue, err := strconv.ParseInt(value, 10, size)
	if err == nil {
		field.SetInt(intValue)
	}

	return err
}

func setUint(field reflect.Value, value string, size int) error {
	if value == "" {
		field.SetUint(0)
		return nil
	}

	uintValue, err := strconv.ParseUint(value, 10, size)
	if err == nil {
		field.SetUint(uintValue)
	}

	return err
}

func setFloat(field reflect.Value, value string, size int) error {
	if value == "" {
		field.SetFloat(0.0)
		return nil
	}

	floatValue, err := strconv.ParseFloat(value, size)
	if err == nil {
		field.SetFloat(floatValue)
	}

	return err
}

func setSlice(field reflect.Value, value []string) error {
	lenValue := len(value)

	sliceKind := field.Type().Elem().Kind()
	slice := reflect.MakeSlice(field.Type(), lenValue, lenValue)

	for i := 0; i < lenValue; i++ {
		if err := setField(sliceKind, slice.Index(i), value[i]); err != nil {
			return err
		}
	}

	field.Set(slice)

	return nil
}
