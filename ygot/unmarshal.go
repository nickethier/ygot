package ygot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/openconfig/gnmi/errlist"
)

func UnmarshalJSON(b []byte, s GoStruct, f JSONFormat) error {
	j := make(map[string]interface{})
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	//fmt.Println(j)
	return constructGoStruct(s, "", j)
}

func constructGoStruct(s GoStruct, parentMod string, j map[string]interface{}) error {
	var errs errlist.List
	sval := reflect.ValueOf(s).Elem()
	stype := sval.Type()
	//fmt.Println(sval)
	//fmt.Println(stype)
	for i := 0; i < sval.NumField(); i++ {
		field := sval.Field(i)

		fType := stype.Field(i)

		//var appmod string
		pmod := parentMod
		if chMod, ok := fType.Tag.Lookup("module"); ok {
			// If the child module isn't the same as the parent module,
			// then appmod stores the name of the module to prefix to paths
			// within this context.
			//	if chMod != parentMod {
			//appmod = chMod
			//	}
			// Update the parent module name to be used for subsequent
			// children.
			pmod = chMod
		}

		constructGoStructValue(field, fType, pmod, j)
		//fmt.Printf("%v -- %v -- %v -- %v -- %v -- %v\n", field, fType, appmod, pmod, *mapPaths[0], errs)
		//fmt.Println(s)
	}

	return errs.Err()
}

func constructGoStructValue(field reflect.Value, fType reflect.StructField, pmod string, j map[string]interface{}) error {

	path, ok := fType.Tag.Lookup("path")

	if !ok {
		return nil
	}

	mapPaths, err := structTagToLibPaths(fType, newStringSliceGNMIPath([]string{}))
	if err != nil {
		return err
	}
	for _, p := range mapPaths {
		for i := 0; i < p.Len(); i++ {
			e, _ := p.StringElemAt(i)
			fmt.Printf("%s -- %s\n", path, e)
		}
	}

	iface := extractInterfaceByPath(j, path)
	if iface == nil {
		return nil
	}
	cVal := reflect.ValueOf(iface)
	switch cVal.Kind() {
	case reflect.Map:
		switch field.Kind() {
		case reflect.Ptr:
			newField := reflect.New(fType.Type.Elem())
			field.Set(newField)
			return constructGoStruct(field.Interface().(GoStruct), pmod, iface.(map[string]interface{}))
		case reflect.Map:
			newField := reflect.MakeMap(fType.Type)
			field.Set(newField)
			for _, k := range cVal.MapKeys() {
				mVal := cVal.MapIndex(k).Elem()
				mType := field.Type().Elem()
				var mapVal reflect.Value
				if mType.Kind() == reflect.Ptr {
					mapVal = reflect.New(mType.Elem())
				} else {
					mapVal = reflect.New(mType)
				}
				err := constructGoStruct(mapVal.Interface().(GoStruct), pmod, mVal.Interface().(map[string]interface{}))
				field.SetMapIndex(k, mapVal)
				if err != nil {
					return err
				}
			}
		}

	case reflect.Slice:

	case reflect.String:
		switch field.Kind() {
		case reflect.Ptr:
			sVal := cVal.Interface().(string)
			field.Set(reflect.ValueOf(&sVal))

		case reflect.String:
			field.Set(cVal)
		default:
			return &json.UnmarshalTypeError{
				Value:  cVal.String(),
				Type:   cVal.Type(),
				Struct: path,
				Field:  fType.Name,
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64:
		vType := field.Type()
		if field.Kind() == reflect.Ptr {
			vType = vType.Elem()
			vVal := reflect.New(field.Type().Elem())
			field.Set(vVal)
		}
		field.Elem().Set(cVal.Convert(vType))

	}
	return nil
}

func extractInterfaceByPath(data map[string]interface{}, path string) (iface interface{}) {
	d := data
	var ok bool
	for _, el := range strings.Split(path, "|") {
		for _, p := range strings.Split(el, "/") {
			iface = d[p]
			if d, ok = d[p].(map[string]interface{}); !ok {
				return
			}
		}
	}
	return
}
