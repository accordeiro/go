package gql

func (_ *resolver) Assets() []*Asset {
	return nil
}

func (_ *resolver) Issuers() []*Issuer {
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
) []*PartialMarket {
	return nil
}

func (_ *resolver) Ticker(
	args struct {
		PairName    *string
		NumHoursAgo *int32
	},
) []*PartialAggregatedMarket {
	return nil
}
