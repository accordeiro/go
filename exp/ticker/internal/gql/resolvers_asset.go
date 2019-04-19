package gql

func (r *resolver) Assets() (assets []*asset, err error) {
	dbAssets, err := r.db.GetAllValidAssets()
	if err != nil {
		return
	}

	for _, dbAsset := range dbAssets {
		assets = append(assets, &asset{
			Code:                        dbAsset.Code,
			IssuerAccount:               dbAsset.IssuerAccount,
			Type:                        dbAsset.Type,
			NumAccounts:                 dbAsset.NumAccounts,
			AuthRequired:                dbAsset.AuthRequired,
			AuthRevocable:               dbAsset.AuthRevocable,
			Amount:                      dbAsset.Amount,
			AssetControlledByDomain:     dbAsset.AssetControlledByDomain,
			AnchorAssetCode:             dbAsset.AnchorAssetCode,
			AnchorAssetType:             dbAsset.AnchorAssetType,
			IsValid:                     dbAsset.IsValid,
			DisplayDecimals:             BigInt(dbAsset.DisplayDecimals),
			Name:                        dbAsset.Name,
			Desc:                        dbAsset.Desc,
			Conditions:                  dbAsset.Conditions,
			IsAssetAnchored:             dbAsset.IsAssetAnchored,
			FixedNumber:                 BigInt(dbAsset.FixedNumber),
			MaxNumber:                   BigInt(dbAsset.MaxNumber),
			IsUnlimited:                 dbAsset.IsUnlimited,
			RedemptionInstructions:      dbAsset.RedemptionInstructions,
			CollateralAddresses:         dbAsset.CollateralAddresses,
			CollateralAddressSignatures: dbAsset.CollateralAddressSignatures,
			Countries:                   dbAsset.Countries,
			Status:                      dbAsset.Status,
			IssuerID:                    dbAsset.IssuerID,
		})
	}

	return
}
