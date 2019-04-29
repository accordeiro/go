package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stellar/go/xdr"
)

type Transaction struct {
	TxEnvelope string `db:"tx_envelope"`
}

type TxInfo struct {
	SendAssetCode   string
	SendAssetIssuer string
	SendMax         int64
	DestAssetCode   string
	DestAssetIssuer string
	DestAmount      int64
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func dbConnect(pgURL string) *sqlx.DB {
	dbInfo, err := pq.ParseURL(pgURL)
	check(err)

	dbconn, err := sqlx.Connect("postgres", dbInfo)
	check(err)
	return dbconn
}

func getTransactionsFromDB(session *sqlx.DB) []Transaction {
	var txs []Transaction
	err := session.Select(
		&txs, `
		SELECT tx_envelope FROM history_transactions htx
		INNER JOIN history_operations hop ON htx.id = hop.transaction_id
		WHERE hop.type = `+fmt.Sprintf("%d", xdr.OperationTypePathPayment),
	)
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

func getAssetCode(asset xdr.Asset) string {
	switch asset.Type {
	case xdr.AssetTypeAssetTypeNative:
		return "XLM"
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		return string(asset.AlphaNum4.AssetCode[:])
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		return string(asset.AlphaNum12.AssetCode[:])
	default:
		return ""
	}
}

func parseTxInfo(txEnvelope xdr.TransactionEnvelope) []TxInfo {
	var txInfos []TxInfo
	for _, op := range txEnvelope.Tx.Operations {
		var txInfo TxInfo
		if op.Body.Type == xdr.OperationTypePathPayment {
			pOp := op.Body.PathPaymentOp

			txInfo.SendMax = int64(pOp.SendMax)
			txInfo.DestAmount = int64(pOp.DestAmount)

			sendAsset := pOp.SendAsset
			txInfo.SendAssetCode = getAssetCode(sendAsset)

			destAsset := pOp.DestAsset
			txInfo.DestAssetCode = getAssetCode(destAsset)

			txInfos = append(txInfos, txInfo)
		}
	}
	return txInfos
}

func main() {
	dbURL := "postgres://stellar:horizon@localhost:8002/horizon?sslmode=disable"
	session := dbConnect(dbURL)
	txs := getTransactionsFromDB(session)

	for _, tx := range txs {
		data := decodeEnvelope(tx.TxEnvelope)
		fmt.Println(parseTxInfo(data))
	}
}
