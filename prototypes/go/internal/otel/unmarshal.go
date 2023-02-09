package otel

import "fmt"

type StringMap map[string]interface{}

func rawToOtlp(raw interface{}) Otlp {
	in := raw.(map[string]interface{})
	insecure := false
	ret := Otlp{
		Insecure: &insecure,
	}
	if val, ok := in["endpoint"]; ok {
		ret.Endpoint = fmt.Sprintf("%v", val)
	}
	if val, ok := in["protocol"]; ok {
		ret.Protocol = fmt.Sprintf("%v", val)
	}
	if val, ok := in["insecure"]; ok {
		insecure = val.(bool)
	}
	return ret
}
