package workflow

import "go.uber.org/cadence/encoded"

type DataConverter struct {
}

func (d DataConverter) ToData(value ...interface{}) ([]byte, error) {
	return encoded.GetDefaultDataConverter().ToData(value)
}

func (d DataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	return encoded.GetDefaultDataConverter().FromData(input, valuePtr...)
}

var _ encoded.DataConverter = (*DataConverter)(nil)
