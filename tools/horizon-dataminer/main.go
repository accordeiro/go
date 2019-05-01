package main

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stellar/go/amount"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/stellar/go/xdr"
)

type Transaction struct {
	TxEnvelope      string    `db:"tx_envelope"`
	LedgerCloseTime time.Time `db:"closed_at"`
}

type TxInfo struct {
	SendAssetCode   string
	SendMax         float64
	DestAssetCode   string
	DestAmount      float64
	LedgerCloseTime time.Time
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func dbConnect(pgURL string) *sqlx.DB {
	dbInfo, err := pq.ParseURL(pgURL)
	check(err)

	dbconn, err := sqlx.Connect("postgres", dbInfo)
	check(err)
	return dbconn
}

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

func retrieveAllPathPayments(session *sqlx.DB, assetCode string, numDaysAgo int) []TxInfo {
	txs := getTransactionsFromDB(session, numDaysAgo)
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

func txToStringSlice(tx TxInfo) []string {
	return []string{
		tx.SendAssetCode,
		fmt.Sprint(tx.SendMax),
		tx.DestAssetCode,
		fmt.Sprint(tx.DestAmount),
		fmt.Sprint(tx.LedgerCloseTime),
	}
}

func writeTxInfosToCSV(txInfos []TxInfo, outFile string) {
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

func nextCursor(nextPageURL string) (cursor string, err error) {
	u, err := url.Parse(nextPageURL)
	if err != nil {
		return
	}

	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return
	}
	cursor = m["cursor"][0]

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

func scrapePathPayments() {
	var pathPayments []TxInfo
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

		pathPayments = append(pathPayments, ppOps...)

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
			if oldestDate.Before(lastDate) {
				break
			}
		}
	}

	writeTxInfosToCSV(pathPayments, "out.csv")
}

func main() {
	scrapePathPayments()

	// dbURL := "postgres://stellar:horizon@localhost:8002/horizon?sslmode=disable"
	// session := dbConnect(dbURL)

	// assetTxs := retrieveAllPathPayments(session, "EUR", 730)
	// writeTxInfosToCSV(assetTxs, "out.csv")
}
