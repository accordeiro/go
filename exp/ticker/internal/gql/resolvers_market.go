package gql

func (r *resolver) Markets(
	args struct {
		BaseAssetCode      *string
		BaseAssetIssuer    *string
		CounterAssetCode   *string
		CounterAssetIssuer *string
		NumHoursAgo        *int32
	},
) ([]*partialMarket, error) {
	var partialMkts []*partialMarket

	numHours := int(96)
	if args.NumHoursAgo != nil {
		numHours = int(*args.NumHoursAgo)
	}

	dbMarkets, err := r.db.RetrievePartialMarkets(
		args.BaseAssetCode,
		args.BaseAssetIssuer,
		args.CounterAssetCode,
		args.CounterAssetIssuer,
		numHours,
	)
	if err != nil {
		return nil, err
	}

	for _, dbMkt := range dbMarkets {
		partialMkts = append(partialMkts, &partialMarket{
			TradePair: dbMkt.TradePairName,
			// TODO: provide code and issuer instead of ids
			BaseAssetID:    dbMkt.BaseAssetID,
			CounterAssetID: dbMkt.CounterAssetID,
			BaseVolume:     dbMkt.BaseVolume,
			CounterVolume:  dbMkt.CounterVolume,
			TradeCount:     dbMkt.TradeCount,
			Open:           dbMkt.Open,
			Low:            dbMkt.Low,
			High:           dbMkt.High,
			Change:         dbMkt.Change,
			Close:          dbMkt.Close,
			// CloseTime: dbMkt.CloseTime,
		})
	}
	return partialMkts, err
}

func (_ *resolver) Ticker(
	args struct {
		PairName    *string
		NumHoursAgo *int32
	},
) []*partialAggregatedMarket {
	return nil
}
