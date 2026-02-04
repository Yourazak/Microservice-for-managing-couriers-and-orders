package usecase

import "time"

type TransportType string

const (
	OnFoot  TransportType = "on_foot"
	Scooter TransportType = "scooter"
	Car     TransportType = "car"
)

type DeliveryTimeFactory struct{}

func NewDeliveryTimeFactory() *DeliveryTimeFactory {
	return &DeliveryTimeFactory{}
}

func (f *DeliveryTimeFactory) Deadline(now time.Time, transport string) time.Time {
	switch transport {
	case string(Scooter):
		return now.Add(15 * time.Minute)
	case string(Car):
		return now.Add(5 * time.Minute)
	default:
		return now.Add(30 * time.Minute)
	}
}
