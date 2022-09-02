package computing

type Computing interface {
	GetID() string
	GetCost() float64
	SetCost(float64)
	GetRegion() string
	GetZone() string
}
