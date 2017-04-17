package main

import (
	"fmt"
)

type Vehicle struct {
	ID       int64   `json:"id"`
	MaxSpeed float64 `json:"maxSpeed"`
	Capacity int64	 `json:"capacity"`
	Route    []int64 `json:"route"`
}

func (v *Vehicle) String() string {
	return fmt.Sprintf("%+v", *v)
}
