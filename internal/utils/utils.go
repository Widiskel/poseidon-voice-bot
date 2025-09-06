package utils

import (
	"encoding/json"
	"reflect"
)

func FormatObject(obj interface{}) (string, error) {
	loggableMap := make(map[string]interface{})

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {

		jsonOutput, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return "", err
		}
		return string(jsonOutput), nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Func {

			loggableMap[fieldType.Name] = "<function>"
			continue
		}

		if field.CanInterface() {
			loggableMap[fieldType.Name] = field.Interface()
		}
	}

	jsonOutput, err := json.MarshalIndent(loggableMap, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonOutput), nil
}
