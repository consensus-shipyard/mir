package parameterset

import (
	"fmt"
)

type Set map[string]any

type Setting struct {
	Key string
	Val string
}

type Element []Setting

func add(i1 Element, i2 Element) Element {
	settings := make([]Setting, len(i1), len(i1)+len(i2))
	copy(settings, i1)
	settings = append(settings, i2...)
	return settings
}

func (s Set) Elements() []Element {
	//var items []Element
	//
	if p, ok := s["product"]; ok {
		return product(p.([]any))
	} else if u, ok := s["union"]; ok {
		return union(u.([]any))
	} else if r, ok := s["range"]; ok {
		name := r.(map[string]any)["name"].(string)
		min := r.(map[string]any)["min"].(int)
		step := r.(map[string]any)["step"].(int)
		max := r.(map[string]any)["max"].(int)
		return rangeFunc(name, min, step, max)
	} else if e, ok := s["enum"]; ok {
		name := e.(map[string]any)["name"].(string)
		vals := e.(map[string]any)["values"].([]any)
		return enum(name, vals)
	} else if e, ok := s["single"]; ok {
		name := e.(map[string]any)["name"].(string)
		val := e.(map[string]any)["value"]
		return enum(name, []any{val})
	} else {
		panic(fmt.Errorf("unknown type of parameter set: %v", s))
	}
}

func enum(name string, values []any) []Element {
	items := make([]Element, 0)
	for _, val := range values {
		items = append(items, []Setting{{
			Key: name,
			Val: val.(string),
		}})
	}
	return items
}

func rangeFunc(name string, val, step, max int) []Element {
	// This function is not called `range` only because it collides with the reserved Go keyword.
	items := make([]Element, 0)
	for ; val <= max; val += step {
		items = append(items, []Setting{{
			Key: name,
			Val: fmt.Sprintf("%d", val),
		}})
	}
	return items
}

func union(sets []any) []Element {
	items := make([]Element, 0)
	for _, set := range sets {
		s := Set(set.(map[string]any))
		items = append(items, s.Elements()...)
	}
	return items
}

func product(sets []any) []Element {
	items := []Element{{}}

	for _, set := range sets {
		s := Set(set.(map[string]any))

		newItems := make([]Element, 0)
		for _, item := range s.Elements() {
			for _, it := range items {
				merged := add(item, it)
				newItems = append(newItems, merged)
			}
		}
		items = newItems

	}

	return items
}
