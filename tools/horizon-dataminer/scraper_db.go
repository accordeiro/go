package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/amount"
	"github.com/stellar/go/xdr"
)

func getTransactionsFromDB(session *sqlx.DB, numDaysAgo int) []Transaction {
	var txs []Transaction
	baseQ := `
		SELECT tx_envelope, hl.closed_at FROM history_transactions htx
		INNER JOIN history_operations hop ON htx.id = hop.transaction_id
		INNER JOIN history_ledgers hl ON htx.ledger_sequence = hl.sequence
		WHERE hl.closed_at > now() - interval '%d days' AND hop.type = %d`

	q := fmt.Sprintf(baseQ, numDaysAgo, xdr.OperationTypePathPayment)

	err := session.Select(&txs, q)
	check(err)

	return txs
}

func decodeEnvelope(b64Envelope string) xdr.TransactionEnvelope {
	rawr := strings.NewReader(b64Envelope)
	b64r := base64.NewDecoder(base64.StdEncoding, rawr)

	var txEnvelope xdr.TransactionEnvelope
	_, err := xdr.Unmarshal(b64r, &txEnvelope)
	check(err)

	return txEnvelope
}

// stripCtlFromUTF8 strips control characters from a string.
// This is particularly useful here since some asset codes here might come
// with a trailing \0000 character.
func stripCtlFromUTF8(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, str)
}

func getAssetCode(asset xdr.Asset) string {
	switch asset.Type {
	case xdr.AssetTypeAssetTypeNative:
		return "XLM"
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		return stripCtlFromUTF8(string(asset.AlphaNum4.AssetCode[:]))
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		return stripCtlFromUTF8(string(asset.AlphaNum12.AssetCode[:]))
	default:
		return ""
	}
}

func parseAmount(amnt xdr.Int64) float64 {
	amntString := amount.String(amnt)
	amntFloat, _ := strconv.ParseFloat(amntString, 64)
	return amntFloat
}

func parseTxInfo(txEnvelope xdr.TransactionEnvelope, closeTime time.Time) []TxInfo {
	var txInfos []TxInfo
	for _, op := range txEnvelope.Tx.Operations {
		var txInfo TxInfo
		if op.Body.Type == xdr.OperationTypePathPayment {
			pOp := op.Body.PathPaymentOp

			txInfo.LedgerCloseTime = closeTime

			txInfo.SendMax = parseAmount(pOp.SendMax)
			txInfo.DestAmount = parseAmount(pOp.DestAmount)

			sendAsset := pOp.SendAsset
			txInfo.SendAssetCode = getAssetCode(sendAsset)

			destAsset := pOp.DestAsset
			txInfo.DestAssetCode = getAssetCode(destAsset)

			txInfos = append(txInfos, txInfo)
		}
	}
	return txInfos
}

func txIncludesAsset(txi TxInfo, assetCode string) bool {
	if txi.SendAssetCode == assetCode || txi.DestAssetCode == assetCode {
		return true
	}

	// Covering a basic case for anchored assets, where you append a T
	// e.g.: USD -> USDT, EUR -> EURT
	anchorAsset := assetCode + "T"
	if txi.SendAssetCode == anchorAsset || txi.DestAssetCode == anchorAsset {
		return true
	}

	return false
}
