package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/evcc-io/evcc/vehicle/polestar"
	"github.com/hasura/go-graphql-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/oauth2"
)

const (
	publicApiURI = "https://pc-api.polestar.com/eu-north-1/mystar-public/"
	publicApiKey = "da2-js63uvc7c5hwpdudt657d5lyou"
)

type apiKeyTransport struct {
	key string
}

var (
	username string
	password string
)

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-api-key", t.key)
	return http.DefaultTransport.RoundTrip(req)
}

func main() {
	flag.StringVar(&username, "username", os.Getenv("POLESTAR_USER"), "Polestar username")
	flag.StringVar(&password, "password", os.Getenv("POLESTAR_PASSWORD"), "Polestar password")
	flag.Parse()
	log := util.NewLogger("polestar")
	id, err := polestar.NewIdentity(log, username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
	httpClient := request.NewClient(log)
	httpClient.Transport = &oauth2.Transport{
		Base:   httpClient.Transport,
		Source: id,
	}
	client := graphql.NewClient(polestar.ApiURIv2, httpClient)
	ctx := context.Background()

	var carsRes GetCarsResponse
	err = client.Query(ctx, &carsRes, nil, graphql.OperationName("getCars"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}

	publicClient := graphql.NewClient(publicApiURI, &http.Client{Transport: &apiKeyTransport{key: publicApiKey}})

	for _, car := range carsRes.GetConsumerCarsV2 {
		var imagesRes GetCarImagesResponse
		err = publicClient.Query(ctx, &imagesRes, map[string]any{
			"pno34":         graphql.String(car.Pno34),
			"structureWeek": graphql.String(car.StructureWeek),
			"modelYear":     graphql.String(car.ModelYear),
			"locale":        graphql.String("no_NO"),
		}, graphql.OperationName("GetCarImages"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
			time.Sleep(10 * time.Second)
			os.Exit(1)
		}

		carInfo.WithLabelValues(car.VIN, car.ModelYear, car.ModelName, car.RegistrationNo).Set(1)
		for _, img := range imagesRes.GetCarImages.Transparent {
			carImage.WithLabelValues(car.VIN, img.URL, fmt.Sprintf("%d", img.Angle), "transparent").Set(1)
		}
		for _, img := range imagesRes.GetCarImages.Opaque {
			carImage.WithLabelValues(car.VIN, img.URL, fmt.Sprintf("%d", img.Angle), "opaque").Set(1)
		}
		fmt.Printf("Found car %s (%s)\n", car.VIN, car.RegistrationNo)
	}

	for _, car := range carsRes.GetConsumerCarsV2 {
		ticker := time.NewTicker(time.Minute)
		quit := make(chan struct{})
		var atomicTelemetry atomic.Pointer[CarTelemetryData]
		atomicTelemetry.Store(&CarTelemetryData{})
		prometheus.MustRegister(NewCollector(car.VIN, &atomicTelemetry))
		go func() {
			fetchTelemetry := func() {
				var tempRes GetCarTelemetryResponse
				err = client.Query(ctx, &tempRes, map[string]any{
					"vins": []string{car.VIN},
				}, graphql.OperationName("CarTelematicsV2"))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Telemetry query for %s failed: %v\n", car.VIN, err)
					return
				}
				atomicTelemetry.Store(&tempRes.CarTelemetryData)
			}
			fetchTelemetry()
			for {
				select {
				case <-ticker.C:
					fetchTelemetry()
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	prometheus.MustRegister(version.NewCollector("polestar_exporter"))
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Http server failed: %v\n", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
}
