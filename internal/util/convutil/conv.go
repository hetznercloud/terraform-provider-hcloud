package convutil

func Int32ToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	result := int(*value)
	return &result
}
