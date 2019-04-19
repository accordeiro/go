package gql

import (
	"log"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stellar/go/exp/ticker/internal/gql/schema"
	"github.com/stellar/go/exp/ticker/internal/tickerdb"
)

type asset tickerdb.Asset

type issuer tickerdb.Issuer

// PartialMarket represents the aggregated market data for a
// specific pair of assets since <Since>
type partialMarket struct {
	BaseAssetID    int32
	CounterAssetID int32
	BaseVolume     float64
	CounterVolume  float64
	TradeCount     int32
	Open           float64
	Low            float64
	High           float64
	Change         float64
	Close          float64
	CloseTime      graphql.Time
	Since          graphql.Time
}

// PartialAggregatedMarket represents the aggregated market data for
// a generic trade pair since <Since>
type partialAggregatedMarket struct {
	TradePair     string
	BaseVolume    float64
	CounterVolume float64
	TradeCount    int32
	Open          float64
	Low           float64
	High          float64
	Change        float64
	Close         float64
	CloseTime     graphql.Time
	Since         graphql.Time
}

type resolver struct{}

func Serve() {
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	s := graphql.MustParseSchema(schema.String(), &resolver{}, opts...)
	http.Handle("/query", &relay.Handler{Schema: s})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
