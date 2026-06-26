package domain

import (
	"strconv"
)

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

type MapHttpParam struct {
	values map[string][]string
}

func NewHttpParam() *mapHttpParam {
	return &mapHttpParam{values: map[string][]string{}}
}

func (p *mapHttpParam) Get(key string) ([]string, bool) {
	values, ok := p.values[key]
	return values, ok
}

func (p *mapHttpParam) GetFirst(key string) (string, bool) {
	values, ok := p.values[key]
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func (p *mapHttpParam) GetFirstOrDefault(key string, defaultValue string) string {
	value, ok := p.GetFirst(key)
	if !ok {
		return defaultValue
	}
	return value
}

func (p *mapHttpParam) GetAll() map[string][]string {
	result := make(map[string][]string, len(p.values))
	for key, values := range p.values {
		result[key] = append([]string(nil), values...)
	}
	return result
}

func (p *mapHttpParam) Set(key string, value string) {
	p.values[key] = []string{value}
}

func (p *mapHttpParam) SetAll(params map[string][]string) {
	p.values = make(map[string][]string, len(params))
	for key, values := range params {
		p.values[key] = append([]string(nil), values...)
	}
}

func (p *mapHttpParam) Delete(key string) {
	delete(p.values, key)
}

func (p *mapHttpParam) GetFirstInt(key string) (int, bool) {
	value, ok := p.GetFirst(key)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func (p *mapHttpParam) GetFirstIntOrDefault(key string, defaultValue int) int {
	value, ok := p.GetFirstInt(key)
	if !ok {
		return defaultValue
	}
	return value
}

func (p *mapHttpParam) GetFirstFloat(key string) (float64, bool) {
	value, ok := p.GetFirst(key)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func (p *mapHttpParam) GetFirstFloatOrDefault(key string, defaultValue float64) float64 {
	value, ok := p.GetFirstFloat(key)
	if !ok {
		return defaultValue
	}
	return value
}

func (p *mapHttpParam) GetFirstBool(key string) (bool, bool) {
	value, ok := p.GetFirst(key)
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}

func (p *mapHttpParam) GetFirstBoolOrDefault(key string, defaultValue bool) bool {
	value, ok := p.GetFirstBool(key)
	if !ok {
		return defaultValue
	}
	return value
}

func (p *mapHttpParam) Clear() {
	p.values = map[string][]string{}
}
