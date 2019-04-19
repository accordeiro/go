package gql

import (
	"github.com/stellar/go/exp/ticker/internal/tickerdb"
)

func (r *resolver) Issuers() ([]*tickerdb.Issuer, error) {
	var issuers []*tickerdb.Issuer
	dbIssuers, err := r.db.GetAllIssuers()
	if err != nil {
		return issuers, err
	}

	for i := range dbIssuers {
		issuers = append(issuers, &dbIssuers[i])
	}

	return issuers, err
}
