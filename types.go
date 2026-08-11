package main

type ConsumerCar struct {
	VIN                        string `graphql:"vin"`
	InternalVehicleIdentifier  string
	Market                     string
	OriginalMarket             string
	Pno34                      string `graphql:"pno34"`
	ModelYear                  string
	ModelName                  string
	StructureWeek              string
	Edition                    string
	RegistrationNo             string
	DeliveryDate               string
	CurrentPlannedDeliveryDate string
	PrimaryDriver              string
}

type CarImage struct {
	URL   string `graphql:"url"`
	Angle int
}

type CarImages struct {
	Transparent []CarImage
	Opaque      []CarImage
}
