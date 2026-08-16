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

type CarTelemetryData struct {
	Health   []HealthData
	Battery  []BatteryData
	Odometer []OdometerData
}

type EventUpdatedTimestamp struct {
	Seconds string `graphql:"seconds"`
	Nanos   int64  `graphql:"nanos"`
}

type HealthData struct {
	VIN                       string `graphql:"vin"`
	BrakeFluidLevelWarning    string
	DaysToService             int
	DistanceToServiceKm       int
	EngineCoolantLevelWarning string
	OilLevelWarning           string
	ServiceWarning            string
	Timestamp                 EventUpdatedTimestamp
}

type BatteryData struct {
	VIN                                string `graphql:"vin"`
	ChargingStatusV2                   string `graphql:"chargingStatusV2"`
	BatteryChargeLevelPercentage       int
	EstimatedChargingTimeToFullMinutes int
	EstimatedDistanceToEmptyKm         int
	Timestamp                          EventUpdatedTimestamp
}

type OdometerData struct {
	VIN            string `graphql:"vin"`
	OdometerMeters int64
	Timestamp      EventUpdatedTimestamp
}
