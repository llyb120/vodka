package vodka

import (
	"vodka/mapper"
)

func ScanMapper(dir string) error {
	return mapper.ScanMapper(dir)
}

func InitMapper(source interface{}) error {
	return mapper.InitMapper(source)
}
