package main

import (
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/stellar/go/xdr"
)

func writeTxInfosToDB(session *sqlx.DB, txInfos []TxInfo) error {
	for _, txInfo := range txInfos {
		_, err := session.NamedExec(
			`INSERT INTO txinfos
			(send_asset_code, send_max, dest_asset_code, dest_amount, ledger_close_time)
			VALUES (:send_asset_code, :send_max, :dest_asset_code, :dest_amount)`,
			txInfo,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func pathPaymentToTxInfo(op operations.PathPayment) (txi TxInfo) {
	txi.LedgerCloseTime = op.Payment.Base.LedgerCloseTime
	txi.SendMax, _ = strconv.ParseFloat(op.SourceMax, 64)
	txi.SendAssetCode = op.SourceAssetCode
	txi.DestAmount, _ = strconv.ParseFloat(op.Payment.Amount, 64)
	txi.DestAssetCode = op.Payment.Asset.Code

	if txi.SendAssetCode == "" {
		txi.SendAssetCode = "XLM"
	}

	if txi.DestAssetCode == "" {
		txi.DestAssetCode = "XLM"
	}

	return
}

func filterPathPaymentOps(ops []operations.Operation) []TxInfo {
	var filteredOps []TxInfo
	for _, op := range ops {
		if op.GetType() == operations.TypeNames[xdr.OperationTypePathPayment] {
			filteredOps = append(
				filteredOps,
				pathPaymentToTxInfo(op.(operations.PathPayment)),
			)
		}
	}

	return filteredOps
}
