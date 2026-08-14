package utils

func CheckError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func CheckErrorm(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}

func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
