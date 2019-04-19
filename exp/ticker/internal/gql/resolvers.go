package gql

func (_ *resolver) Issuers() []*issuer {
	return nil
}

func (_ *resolver) Markets(
	args struct {
		BaseAssetCode      *string
		BaseAssetIssuer    *string
		CounterAssetCode   *string
		CounterAssetIssuer *string
		NumHoursAgo        *int32
	},
) []*partialMarket {
	return nil
}

func (_ *resolver) Ticker(
	args struct {
		PairName    *string
		NumHoursAgo *int32
	},
) []*partialAggregatedMarket {
	return nil
}
