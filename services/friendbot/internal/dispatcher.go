package internal

import (
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/errors"
)

// MinionDispatcher ...
type MinionDispatcher interface {
	SubmitTransaction(minion *Minion, hclient *horizonclient.Client, tx string) (*hProtocol.Transaction, error)
	CheckSequenceRefresh(minion *Minion, hclient *horizonclient.Client) error
}

// BasicMinionDispatcher ...
type BasicMinionDispatcher struct{}

// SubmitTransaction should be passed to the Minion.
func (m *BasicMinionDispatcher) SubmitTransaction(minion *Minion, hclient *horizonclient.Client, tx string) (*hProtocol.Transaction, error) {
	return submitTransaction(minion, hclient, tx)
}

func submitTransaction(minion *Minion, hclient *horizonclient.Client, tx string) (*hProtocol.Transaction, error) {
	result, err := hclient.SubmitTransactionXDR(tx)
	if err != nil {
		errStr := "submitting tx to horizon"
		switch e := err.(type) {
		case *horizonclient.Error:
			minion.checkHandleBadSequence(e)
			resStr, resErr := e.ResultString()
			if resErr != nil {
				errStr += ": error getting horizon error code: " + resErr.Error()
			} else if resStr == createAccountAlreadyExistXDR {
				return nil, errors.Wrap(ErrAccountExists, errStr)
			} else {
				errStr += ": horizon error string: " + resStr
			}
			return nil, errors.New(errStr)
		}
		return nil, errors.Wrap(err, errStr)
	}
	return &result, nil
}

// CheckSequenceRefresh establishes the minion's initial sequence number, if needed.
// This should also be passed to the minion.
func (m *BasicMinionDispatcher) CheckSequenceRefresh(minion *Minion, hclient *horizonclient.Client) error {
	return checkSequenceRefresh(minion, hclient)
}

func checkSequenceRefresh(minion *Minion, hclient *horizonclient.Client) error {
	if minion.Account.Sequence != 0 && !minion.forceRefreshSequence {
		return nil
	}
	err := minion.Account.RefreshSequenceNumber(hclient)
	if err != nil {
		return errors.Wrap(err, "refreshing minion seqnum")
	}
	minion.forceRefreshSequence = false
	return nil
}
