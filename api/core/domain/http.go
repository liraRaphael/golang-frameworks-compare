package domain

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ParameterIn string

const (
	ParameterInPath   ParameterIn = "path"
	ParameterInQuery  ParameterIn = "query"
	ParameterInHeader ParameterIn = "header"
	ParameterInCookie ParameterIn = "cookie"
)

func (p ParameterIn) String() string {
	return string(p)
}

type HttpParamsType map[string]HttpParamType

type HttpParamType []string

type HttpParam interface {
	Get(key string) ([]string, bool)
	GetFirst(key string) (string, bool)
	GetFirstOrDefault(key string, defaultValue string) string
	GetAll() map[string][]string
	Set(key string, value string)
	SetAll(params map[string][]string)
	Delete(key string)
	GetFirstInt(key string) (int, bool)
	GetFirstIntOrDefault(key string, defaultValue int) int
	GetFirstFloat(key string) (float64, bool)
	GetFirstFloatOrDefault(key string, defaultValue float64) float64
	GetFirstBool(key string) (bool, bool)
	GetFirstBoolOrDefault(key string, defaultValue bool) bool
	Clear()
}

func NewHttpParam() *HttpParamType {
	return &HttpParamType{}
}

func (p HttpParamType) Get(index int) (string, bool) {
	if index < 0 || index >= len(p) {
		return "", false
	}

	return p[index], true

}

func (p *HttpParamType) GetFirst() (string, bool) {
	if len(*p) == 0 {
		return "", false
	}
	return (*p)[0], true
}

func (p *HttpParamType) GetFirstOrDefault(defaultValue string) string {
	value, ok := p.GetFirst()
	if !ok {
		return defaultValue
	}
	return value
}

func (p *HttpParamType) GetAll() []string {
	return []string(*p)
}

func (p *HttpParamType) Set(value string) {
	l := append(*p, value)
	p = &l
}

func (p *HttpParamType) SetAll(params []string) {
	l := append(*p, params...)
	p = &l
}

func (p *HttpParamType) GetFirstInt() (int, bool) {
	value, ok := p.GetFirst()
	if !ok {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func (p *HttpParamType) GetFirstIntOrDefault(defaultValue int) int {
	value, ok := p.GetFirstInt()
	if !ok {
		return defaultValue
	}
	return value
}

func (p *HttpParamType) GetFirstFloat() (float64, bool) {
	value, ok := p.GetFirst()
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func (p *HttpParamType) GetFirstFloatOrDefault(defaultValue float64) float64 {
	value, ok := p.GetFirstFloat()
	if !ok {
		return defaultValue
	}
	return value
}

func (p *HttpParamType) GetFirstBool() (bool, bool) {
	value, ok := p.GetFirst()
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}

func (p *HttpParamType) GetFirstBoolOrDefault(defaultValue bool) bool {
	value, ok := p.GetFirstBool()
	if !ok {
		return defaultValue
	}
	return value
}

func (p *HttpParamType) Clear() {
	p = &HttpParamType{}
}

func ExtractParamsFromTag(typ reflect.Type) (headersFields []int, queryFields []int, pathFields []int, cookieFields []int) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		inType, _, hasTag := ParseParamTag(field)
		if !hasTag {
			continue
		}

		switch inType {
		case ParameterInHeader:
			headersFields = append(headersFields, i)
		case ParameterInQuery:
			queryFields = append(queryFields, i)
		case ParameterInPath:
			pathFields = append(pathFields, i)
		case ParameterInCookie:
			cookieFields = append(cookieFields, i)
		default:
			continue
		}
	}
	return

}

func ParamsValuesToHttpParamsType(vOf reflect.Value, indexs []int) HttpParamsType {
	typ := vOf.Type()
	kind := vOf.Kind()

	v := HttpParamsType{}
	for _, field := range indexs {
		if IsNilValue(vOf) {
			continue
		}

		_, tagValue, _ := ParseParamTag(typ.Field(field))

		switch kind {
		case reflect.Slice, reflect.Array:
			for i := 0; i < vOf.Len(); i++ {
				v[tagValue] = append(v[tagValue], fmt.Sprintf("%v", vOf.Index(i).Field(field).Interface()))
			}
		default:
			v[tagValue] = []string{fmt.Sprintf("%v", vOf.Field(field).Interface())}
		}
	}

	return v
}

func SetFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field is not settable")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}

	return nil
}

func IsNilValue(value any) bool {
	if value == nil {
		return true
	}

	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflectValue.IsNil()
	default:
		return false
	}
}

func (params HttpParamsType) GetValues(name string) []string {
	if params == nil {
		return nil
	}
	values, ok := params[name]
	if !ok {
		return nil
	}
	return values
}

func ParseParamTag(field reflect.StructField) (ParameterIn, string, bool) {
	tags := []struct {
		inType ParameterIn
		tag    string
	}{
		{ParameterInQuery, field.Tag.Get(ParameterInQuery.String())},
		{ParameterInHeader, field.Tag.Get(ParameterInHeader.String())},
		{ParameterInPath, field.Tag.Get(ParameterInPath.String())},
		{ParameterInCookie, field.Tag.Get(ParameterInCookie.String())},
	}

	for _, entry := range tags {
		if entry.tag == "" {
			continue
		}
		parts := strings.Split(entry.tag, ",")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		return entry.inType, parts[0], true
	}

	return "", "", false
}

func ValidPointerType[ParamType any](params ParamType) (reflect.Value, error) {
	value := reflect.ValueOf(params)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return value, fmt.Errorf("Empty type parameter")
	}

	elem := value.Elem()
	if !elem.CanSet() {
		return value, fmt.Errorf("target must be settable")
	}

	if elem.Kind() != reflect.Struct {
		return value, fmt.Errorf("target must be a struct")
	}

	return value, nil
}

func PopulateParams[ParamType any](headers, queryParams, pathParams, cookies HttpParamsType) (ParamType, error) {
	var params ParamType
	elem, err := ValidPointerType(params)
	if err != nil {
		return params, err
	}

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elem.Type().Field(i)
		if !field.CanSet() {
			continue
		}

		inType, name, hasTag := ParseParamTag(fieldType)
		if !hasTag {
			continue
		}

		var values []string
		switch inType {
		case ParameterInHeader:
			values = headers.GetValues(name)
		case ParameterInQuery:
			values = queryParams.GetValues(name)
		case ParameterInPath:
			values = pathParams.GetValues(name)
		case ParameterInCookie:
			values = cookies.GetValues(name)
		default:
			continue
		}

		if len(values) == 0 {
			continue
		}

		if err := SetFieldValue(field, values[0]); err != nil {
			return params, err
		}
	}

	return params, nil
}
