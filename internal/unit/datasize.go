package unit

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

type UnitValue struct {
	Unit  string  `json:"unit"`
	Value float32 `json:"value"`
}

func (uv *UnitValue) String() string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", uv.Value), "0"), ".") + " " + uv.Unit
}

const (
	// base 10 (SI prefixes)
	Bit     Data = 1e0
	Kilobit Data = 1e3
	Megabit Data = 1e6
	Gigabit Data = 1e9
)

var mapUnit = map[string]Data{
	"B":  Bit,
	"KB": Kilobit,
	"MB": Megabit,
	"GB": Gigabit,
}

type Data float64

func Quota(used UnitValue, capacity UnitValue) (float64, error) {
	u, err := From(used)
	if err != nil {
		return 0, err
	}

	c, err := From(capacity)
	if err != nil {
		return 0, err
	}

	return u.Bits() / c.Bits(), nil
}

func From(uv UnitValue) (Data, error) {
	unit, ok := mapUnit[uv.Unit]
	if !ok {
		return unit, xerrors.Errorf("unsupported unit %s", uv.Unit)
	}
	return unit, nil
}

func (b Data) Bits() float64 {
	return float64(b)
}

func (b Data) Kilobits() float64 {
	return float64(b / Kilobit)
}

func (b Data) Megabits() float64 {
	return float64(b / Megabit)
}

func (b Data) Gigabits() float64 {
	return float64(b / Gigabit)
}
