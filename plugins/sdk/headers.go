package sdk

import "strings"

func (r HostResponse) FirstHeader(name string) string {
	values := r.HeaderValuesFor(name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (r HostResponse) HeaderValuesFor(name string) []string {
	if values, ok := r.HeaderValues[name]; ok {
		return values
	}
	for key, values := range r.HeaderValues {
		if strings.EqualFold(key, name) {
			return values
		}
	}
	return nil
}
