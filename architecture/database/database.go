package database

type Database interface {
	GetID() string
	GetAvailabilityRate() int
	GetCost() float64
	SetCost(float64)
}
