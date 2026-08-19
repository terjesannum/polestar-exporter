package main

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	carInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "polestar_info",
		Help: "Information about the car",
	}, []string{"vin", "model_year", "model_name", "registration_no"})
	carImage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "polestar_image",
		Help: "Car image",
	}, []string{"vin", "url", "angle", "type"})
)

type TelemetryCollector struct {
	telemetry              *atomic.Pointer[CarTelemetryData]
	serviceSeconds         *prometheus.Desc
	serviceMeters          *prometheus.Desc
	odometerMeters         *prometheus.Desc
	chargingStatus         *prometheus.Desc
	chargingTimeToFull     *prometheus.Desc
	batteryLevel           *prometheus.Desc
	batteryDistanceToEmpty *prometheus.Desc
	healthTimestamp        *prometheus.Desc
	batteryTimestamp       *prometheus.Desc
	odometerTimestamp      *prometheus.Desc
}

func NewCollector(vin string, telemetry *atomic.Pointer[CarTelemetryData]) *TelemetryCollector {
	labels := prometheus.Labels{"vin": vin}
	return &TelemetryCollector{
		telemetry: telemetry,
		serviceSeconds: prometheus.NewDesc(
			"polestar_service_remaining_seconds",
			"Number of seconds until the next service",
			nil,
			labels),
		serviceMeters: prometheus.NewDesc(
			"polestar_service_remaining_meters",
			"Distance in meters until the next required service",
			nil,
			labels),
		odometerMeters: prometheus.NewDesc(
			"polestar_odometer_meters_total",
			"Odometer reading in meters",
			nil,
			labels),
		chargingStatus: prometheus.NewDesc(
			"polestar_charging",
			"Charging status of the car (1 = charging, 0 = not charging)",
			nil,
			labels),
		chargingTimeToFull: prometheus.NewDesc(
			"polestar_charging_time_to_full_seconds",
			"Estimated time to full charge in seconds",
			nil,
			labels),
		batteryLevel: prometheus.NewDesc(
			"polestar_battery_level_percent",
			"Battery charge level in percent",
			nil,
			labels),
		batteryDistanceToEmpty: prometheus.NewDesc(
			"polestar_battery_distance_to_empty_meters",
			"Estimated distance to empty in meters",
			nil,
			labels),
		healthTimestamp: prometheus.NewDesc(
			"polestar_health_timestamp_seconds",
			"Timestamp of the health data",
			nil,
			labels),
		batteryTimestamp: prometheus.NewDesc(
			"polestar_battery_timestamp_seconds",
			"Timestamp of the battery data",
			nil,
			labels),
		odometerTimestamp: prometheus.NewDesc(
			"polestar_odometer_timestamp_seconds",
			"Timestamp of the odometer data",
			nil,
			labels),
	}
}

func (c *TelemetryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.serviceSeconds
	ch <- c.serviceMeters
	ch <- c.odometerMeters
	ch <- c.chargingStatus
	ch <- c.chargingTimeToFull
	ch <- c.batteryLevel
	ch <- c.batteryDistanceToEmpty
	ch <- c.healthTimestamp
	ch <- c.batteryTimestamp
	ch <- c.odometerTimestamp
}

func metricTimestamp(ts EventUpdatedTimestamp) (time.Time, error) {
	seconds, err := strconv.ParseInt(ts.Seconds, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid timestamp: %v\n", ts)
		return time.Time{}, err
	}
	return time.Unix(seconds, ts.Nanos), nil
}

func (c *TelemetryCollector) Collect(ch chan<- prometheus.Metric) {
	data := c.telemetry.Load()
	if data == nil {
		return
	}

	if len(data.Health) == 1 {
		var ts, err = metricTimestamp(data.Health[0].Timestamp)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				c.healthTimestamp,
				prometheus.GaugeValue,
				float64(ts.Unix()))
			ch <- prometheus.MustNewConstMetric(
				c.serviceSeconds,
				prometheus.GaugeValue,
				float64(data.Health[0].DaysToService*24*60*60))
			ch <- prometheus.MustNewConstMetric(
				c.serviceMeters,
				prometheus.GaugeValue,
				float64(data.Health[0].DistanceToServiceKm*1000))
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected health data: %v\n", data.Health)
	}

	if len(data.Battery) == 1 {
		var ts, err = metricTimestamp(data.Battery[0].Timestamp)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				c.batteryTimestamp,
				prometheus.GaugeValue,
				float64(ts.Unix()))
			var charging float64
			if data.Battery[0].ChargingStatusV2 == "CHARGING_STATUS_V2_CHARGING" {
				charging = 1
			} else {
				charging = 0
			}
			ch <- prometheus.MustNewConstMetric(
				c.chargingStatus,
				prometheus.GaugeValue,
				charging)
			ch <- prometheus.MustNewConstMetric(
				c.chargingTimeToFull,
				prometheus.GaugeValue,
				float64(data.Battery[0].EstimatedChargingTimeToFullMinutes*60))
			ch <- prometheus.MustNewConstMetric(
				c.batteryLevel,
				prometheus.GaugeValue,
				float64(data.Battery[0].BatteryChargeLevelPercentage))
			ch <- prometheus.MustNewConstMetric(
				c.batteryDistanceToEmpty,
				prometheus.GaugeValue,
				float64(data.Battery[0].EstimatedDistanceToEmptyKm*1000))
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected battery data: %v\n", data.Health)
	}

	if len(data.Odometer) == 1 {
		var ts, err = metricTimestamp(data.Odometer[0].Timestamp)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				c.odometerTimestamp,
				prometheus.GaugeValue,
				float64(ts.Unix()))
			ch <- prometheus.MustNewConstMetric(
				c.odometerMeters,
				prometheus.CounterValue,
				float64(data.Odometer[0].OdometerMeters))
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected odometer data: %v\n", data.Health)
	}
	}

}
