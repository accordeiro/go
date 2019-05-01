package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stellar/go/clients/horizonclient"
)

type Transaction struct {
	TxEnvelope      string    `db:"tx_envelope"`
	LedgerCloseTime time.Time `db:"closed_at"`
}

type TxInfo struct {
	SendAssetCode   string    `db:"send_asset_code"`
	SendMax         float64   `db:"send_max"`
	DestAssetCode   string    `db:"dest_asset_code"`
	DestAmount      float64   `db:"dest_amount"`
	LedgerCloseTime time.Time `db:"ledger_close_time"`
}

func WriteTxInfosToCSV(txInfos []TxInfo, outFile string) {
	file, err := os.Create(outFile)
	check(err)
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	err = w.Write([]string{"send_asset", "send_max", "dest_asset", "dest_amount", "close_time"})
	check(err)

	for _, tx := range txInfos {
		err = w.Write(txToStringSlice(tx))
		check(err)
	}
}

func RetrieveAllPathPaymentsFromHorizon(session *sqlx.DB) {
	client := horizonclient.DefaultPublicNetClient
	opReq := horizonclient.OperationRequest{
		Order: horizonclient.OrderDesc,
		Limit: 200,
	}
	lastDate := time.Now().AddDate(0, 0, -5)

	opsPage, err := client.Payments(opReq)
	check(err)

	for opsPage.Links.Next.Href != opsPage.Links.Self.Href {
		ppOps := filterPathPaymentOps(opsPage.Embedded.Records)

		err := writeTxInfosToDB(session, ppOps)
		if err != nil {
			fmt.Println("could no insert tx infos:", err)
			continue
		}

		// Finding next page's params:
		nextURL := opsPage.Links.Next.Href
		n, err := nextCursor(nextURL)
		check(err)

		fmt.Println("Cursor currently at:", n)
		opReq.Cursor = n

		opsPage, err = client.Payments(opReq)
		check(err)

		if len(ppOps) > 0 {
			oldestDate := ppOps[len(ppOps)-1].LedgerCloseTime
			fmt.Println("Oldest date ingested:", oldestDate)
			if oldestDate.Before(lastDate) {
				break
			}
		}
	}
}

func RetrieveAllPathPaymentsFromHorizonDB(hrzSession *sqlx.DB, assetCode string, numDaysAgo int) []TxInfo {
	txs := getTransactionsFromDB(hrzSession, numDaysAgo)
	var assetTxs []TxInfo

	// transactions are XDR-encoded, so we can't filter transactions
	// for a specific asset directly in the database query.
	for _, tx := range txs {
		data := decodeEnvelope(tx.TxEnvelope)
		txInfos := parseTxInfo(data, tx.LedgerCloseTime)

		for _, txi := range txInfos {
			if txIncludesAsset(txi, assetCode) {
				assetTxs = append(assetTxs, txi)
			}
		}
	}

	return assetTxs
}

func main() {
	dbURL := "postgres://localhost/dataminer01?sslmode=disable"
	session := dbConnect(dbURL)

	RetrieveAllPathPaymentsFromHorizon(session)

	// assetTxs := retrieveAllPathPayments(session, "EUR", 730)
	// writeTxInfosToCSV(assetTxs, "out.csv")
}
