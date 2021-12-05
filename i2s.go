package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	kindType := reflect.TypeOf(out).Kind()
	if kindType != reflect.Ptr {
		return fmt.Errorf("%v not pointer value", out)
	}
	elemType := reflect.TypeOf(out).Elem().Kind()
	switch elemType {
	case reflect.Struct:
		if err := UnmarshalStruct(data, out); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		if err := UnmarshalSlice(data, out); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported out value type %v", elemType)
	}
	return nil
}

func UnmarshalStruct(data interface{}, out interface{}) error {
	dataKindType := reflect.TypeOf(data).Kind()
	if dataKindType != reflect.Map {
		return fmt.Errorf("invalid data type got: %v, want: %v", dataKindType, reflect.Map)
	}
	value := reflect.ValueOf(data)
	elem := reflect.ValueOf(out).Elem()
	for i := 0; i < elem.NumField(); i++ {
		fieldName := elem.Type().Field(i).Name
		field := elem.Field(i)
		key, ok := getIndex(value.MapKeys(), fieldName)
		if !ok {
			continue
		}
		keyValue := value.MapIndex(key)
		keyType := keyValue.Elem().Kind()
		switch field.Kind() {
		case reflect.Int:
			if keyType != reflect.Float64 {
				return fmt.Errorf("invalid value type got: %v, want: %v", keyType, reflect.Float64)
			}
			floatValue := keyValue.Elem().Float()
			field.SetInt(int64(floatValue))
		case reflect.String:
			if keyType != reflect.String {
				return fmt.Errorf("invalid value type got: %v, want: %v", keyType, reflect.String)
			}
			field.SetString(keyValue.Elem().String())
		case reflect.Bool:
			if keyType != reflect.Bool {
				return fmt.Errorf("invalid value type got: %v, want: %v", keyType, reflect.Bool)
			}
			field.SetBool(keyValue.Elem().Bool())
		case reflect.Array, reflect.Slice:
			if keyType != reflect.Slice {
				return fmt.Errorf("invalid value type got: %v, want: %v", keyType, reflect.Slice)
			}
			if err := i2s(value.MapIndex(key).Interface(), field.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Struct:
			if keyType != reflect.Map {
				return fmt.Errorf("invalid value type got: %v, want: %v", keyType, reflect.Map)
			}
			if err := i2s(keyValue.Interface(), field.Addr().Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func UnmarshalSlice(data interface{}, out interface{}) error {
	dataKindType := reflect.TypeOf(data).Kind()
	if dataKindType != reflect.Slice {
		return fmt.Errorf("bad data structure got: %v, want: %v", dataKindType, reflect.Slice)
	}
	elem := reflect.ValueOf(out).Elem()
	value := reflect.ValueOf(data)
	for i := 0; i < value.Len(); i++ {
		elemType := reflect.TypeOf(out).Elem().Elem()
		elem.Set(reflect.Append(elem, reflect.Zero(elemType)))
		if err := i2s(value.Index(i).Interface(), elem.Index(i).Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

func getIndex(keys []reflect.Value, key string) (reflect.Value, bool) {
	for _, v := range keys {
		if key == v.String() {
			return v, true
		}
	}
	return reflect.Value{}, false
}
