package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/evcc-io/evcc/vehicle/polestar"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

const (
	publicApiURI = "https://pc-api.polestar.com/eu-north-1/mystar-public/"
	publicApiKey = "da2-js63uvc7c5hwpdudt657d5lyou"
)

type apiKeyTransport struct {
	key string
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-api-key", t.key)
	return http.DefaultTransport.RoundTrip(req)
}

func main() {
	log := util.NewLogger("polestar")
	id, err := polestar.NewIdentity(log, os.Getenv("POLESTAR_USER"), os.Getenv("POLESTAR_PASSWORD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
		os.Exit(1)
	}
	httpClient := request.NewClient(log)
	httpClient.Transport = &oauth2.Transport{
		Base:   httpClient.Transport,
		Source: id,
	}
	client := graphql.NewClient(polestar.ApiURIv2, httpClient)
	var carsRes struct {
		GetConsumerCarsV2 []ConsumerCar `graphql:"getConsumerCarsV2"`
	}
	ctx := context.Background()

	err = client.Query(ctx, &carsRes, nil, graphql.OperationName("getCars"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "query failed: %v\n", err)
	}
	fmt.Printf("cars: %v\n\n", carsRes)

	publicClient := graphql.NewClient(publicApiURI, &http.Client{Transport: &apiKeyTransport{key: publicApiKey}})

	for _, car := range carsRes.GetConsumerCarsV2 {
		var imagesRes struct {
			GetCarImages CarImages `graphql:"getCarImages(pno34: $pno34, structureWeek: $structureWeek, modelYear: $modelYear, locale: $locale)"`
		}
		err = publicClient.Query(ctx, &imagesRes, map[string]any{
			"pno34":         graphql.String(car.Pno34),
			"structureWeek": graphql.String(car.StructureWeek),
			"modelYear":     graphql.String(car.ModelYear),
			"locale":        graphql.String("no_NO"),
		}, graphql.OperationName("GetCarImages"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "query failed: %v\n", err)
		}
		fmt.Printf("images: %v\n\n", imagesRes)
	}

	for _, car := range carsRes.GetConsumerCarsV2 {
		var telemetryRes struct {
			CarTelemetryData `graphql:"carTelematicsV2(vins: $vins)"`
		}

		err = client.Query(ctx, &telemetryRes, map[string]any{
			"vins": []string{car.VIN},
		}, graphql.OperationName("CarTelematicsV2"))

		fmt.Printf("telemetry: %v\n\n", telemetryRes)
	}

}
