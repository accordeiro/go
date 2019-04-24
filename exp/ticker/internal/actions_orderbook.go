package ticker

import (
	"fmt"

	horizonclient "github.com/stellar/go/exp/clients/horizon"
	"github.com/stellar/go/exp/ticker/internal/scraper"
	"github.com/stellar/go/exp/ticker/internal/tickerdb"
	"github.com/stellar/go/support/errors"
	hlog "github.com/stellar/go/support/log"
)

// UpdateOrderbookEntries updates the orderbook entries for the relevant markets that were active
// in the past 7-day interval
func UpdateOrderbookEntries(s *tickerdb.TickerSession, c *horizonclient.Client, l *hlog.Entry) error {
	sc := scraper.ScraperConfig{
		Client: c,
		Logger: l,
	}

	mkts, err := s.RetrievePartialMarkets(nil, nil, nil, nil, 168)
	if err != nil {
		return errors.Wrap(err, "could not retrieve partial markets")
	}

	var orderbooks []scraper.OrderbookStats
	for _, mkt := range mkts {
		ob, err := sc.FetchOrderbookForAssets(
			mkt.BaseAssetType,
			mkt.BaseAssetCode,
			mkt.BaseAssetIssuer,
			mkt.CounterAssetType,
			mkt.CounterAssetCode,
			mkt.CounterAssetIssuer,
		)

		if err != nil {
			return errors.Wrap(err, "could not fetch orderbook for assets")
		}

		orderbooks = append(orderbooks, ob)
	}

	fmt.Println(orderbooks)
	return nil
}
