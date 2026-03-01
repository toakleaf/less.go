package less_go

func normalizeDumpLineNumbersOption(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case bool:
		if !v {
			return nil, false
		}
		return "comments", true
	case string:
		if v == "" {
			return nil, false
		}
		return v, true
	default:
		return value, true
	}
}

func dumpLineNumbersEnabled(value any) bool {
	_, enabled := normalizeDumpLineNumbersOption(value)
	return enabled
}
