package utils

import "fmt"

type OrderedMap struct {
	keys  []string
	store map[string]interface{}
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		store: make(map[string]interface{}),
	}
}

func (om *OrderedMap) Keys() []string {
	return om.keys
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, ok := om.store[key]; !ok {
		om.keys = append(om.keys, key)
	}
	om.store[key] = value
}

func (om *OrderedMap) ByKey(key string) interface{} {
	if _, ok := om.store[key]; !ok {
		return nil
	}

	return om.store[key]
}

func (om *OrderedMap) ByIndex(index int) interface{} {
	if index < 0 || index >= len(om.keys) {
		return nil
	}

	return om.store[om.keys[index]]
}

func (om *OrderedMap) String() string {
	var str string
	for i, key := range om.keys {
		str += fmt.Sprintf("%d. %s\n%+v\n", i, key, om.store[key])
	}

	return str
}
