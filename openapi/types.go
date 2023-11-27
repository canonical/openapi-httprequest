package openapi

// TypeString translates an OpenAPI type and translates it to a Go type
func TypeString(typ string, format string) string {
	switch typ {
	case "integer":
		if format == "int32" {
			return "int"
		}
		return "int64"
	case "number":
		return "float64"
	case "string":
		if format == "date-time" || format == "date" {
			return "*time.Time"
		}
		return "string"
	case "boolean":
		return "bool"
	case "object":
		switch format {
		case "map[string]string":
			return format
		default:
			return "interface{}"
		}
	default:
		return "interface{}"
	}
}
