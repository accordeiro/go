package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

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

func txToStringSlice(tx TxInfo) []string {
	return []string{
		tx.SendAssetCode,
		fmt.Sprint(tx.SendMax),
		tx.DestAssetCode,
		fmt.Sprint(tx.DestAmount),
		fmt.Sprint(tx.LedgerCloseTime),
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
