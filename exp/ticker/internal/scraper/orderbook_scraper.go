package scraper

import (
	"math"
	"strconv"

	"github.com/pkg/errors"
	horizonclient "github.com/stellar/go/exp/clients/horizon"
	hProtocol "github.com/stellar/go/protocols/horizon"
)

// FetchOrderbookForAssets fetches the orderbook stats for the base and counter assets provided in the parameters
func (c *ScraperConfig) FetchOrderbookForAssets(bType, bCode, bIssuer, cType, cCode, cIssuer string) (OrderbookStats, error) {
	obStats := OrderbookStats{
		BaseAssetCode:      bType,
		BaseAssetType:      bCode,
		BaseAssetIssuer:    bIssuer,
		CounterAssetCode:   cType,
		CounterAssetType:   cCode,
		CounterAssetIssuer: cIssuer,
		BidMax:             math.Inf(-1), // start with -Inf to make sure we catch the correct max bid
		AskMin:             math.Inf(1),  // start with +Inf to make sure we catch the correct min ask
	}
	r, err := createOrderbookRequest(bType, bCode, bIssuer, cType, cCode, cIssuer)
	if err != nil {
		return obStats, errors.Wrap(err, "could not create a orderbook request")
	}
	summary, err := c.Client.OrderBook(r)
	if err != nil {
		return obStats, errors.Wrap(err, "could not fetch orderbook summary")
	}

	err = calcOrderbookStats(&obStats, summary)
	return obStats, errors.Wrap(err, "could not calculate orderbook stats")
}

// calcOrderbookStats calculates the NumBids, BidVolume, BidMax, NumAsks, AskVolume and AskMin
// statistics for a given OrdebookStats instance
func calcOrderbookStats(obStats *OrderbookStats, summary hProtocol.OrderBookSummary) error {
	obStats.NumBids = len(summary.Bids)
	if obStats.NumBids == 0 {
		obStats.BidMax = 0
	}
	for _, bid := range summary.Bids {
		pricef, err := strconv.ParseFloat(bid.Price, 64)
		if err != nil {
			return errors.Wrap(err, "invalid bid price")
		}
		obStats.BidVolume += pricef
		if pricef > obStats.BidMax {
			obStats.BidMax = pricef
		}
	}

	obStats.NumAsks = len(summary.Asks)
	if obStats.NumAsks == 0 {
		obStats.AskMin = 0
	}
	for _, ask := range summary.Asks {
		pricef, err := strconv.ParseFloat(ask.Price, 64)
		if err != nil {
			return errors.Wrap(err, "invalid ask price")
		}
		obStats.AskVolume += pricef
		if pricef < obStats.AskMin {
			obStats.AskMin = pricef
		}
	}

	obStats.Spread, obStats.SpreadMidPoint = calcSpread(obStats.BidMax, obStats.AskMin)
	return nil
}

// calcSpread calculates the spread stats for the given bidMax and askMin orderbook values
func calcSpread(bidMax float64, askMin float64) (spread float64, midPoint float64) {
	if askMin == 0 || bidMax == 0 {
		return 0, 0
	}
	spread = math.Abs(askMin-bidMax) / askMin
	midPoint = bidMax + spread/2.0
	return
}

// createOrderbookRequest generates a horizonclient.OrderBookRequest based on the base
// and counter asset parameters provided
func createOrderbookRequest(bType, bCode, bIssuer, cType, cCode, cIssuer string) (horizonclient.OrderBookRequest, error) {
	r := horizonclient.OrderBookRequest{}

	switch bType {
	case string(horizonclient.AssetTypeNative):
		r.SellingAssetType = horizonclient.AssetTypeNative
	case string(horizonclient.AssetType4):
		r.SellingAssetType = horizonclient.AssetType4
		r.SellingAssetCode = bCode
		r.SellingAssetIssuer = bIssuer
	case string(horizonclient.AssetType12):
		r.SellingAssetType = horizonclient.AssetType12
		r.SellingAssetCode = bCode
		r.SellingAssetIssuer = bIssuer
	default:
		return r, errors.New("invalid base asset type")
	}

	switch cType {
	case string(horizonclient.AssetTypeNative):
		r.BuyingAssetType = horizonclient.AssetTypeNative
	case string(horizonclient.AssetType4):
		r.BuyingAssetType = horizonclient.AssetType4
		r.BuyingAssetCode = cCode
		r.BuyingAssetIssuer = cIssuer
	case string(horizonclient.AssetType12):
		r.BuyingAssetType = horizonclient.AssetType12
		r.BuyingAssetCode = cCode
		r.BuyingAssetIssuer = cIssuer
	default:
		return r, errors.New("invalid counter asset type")
	}

	return r, nil
}
