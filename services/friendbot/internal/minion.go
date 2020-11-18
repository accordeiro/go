package internal

import (
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
)

const createAccountAlreadyExistXDR = "AAAAAAAAAGT/////AAAAAQAAAAAAAAAA/////AAAAAA="

var ErrAccountExists error = errors.New(fmt.Sprintf("createAccountAlreadyExist (%s)", createAccountAlreadyExistXDR))

// Minion contains a Stellar channel account and Go channels to communicate with friendbot.
type Minion struct {
	Account         Account
	Keypair         *keypair.Full
	BotAccount      txnbuild.Account
	BotKeypair      *keypair.Full
	Horizon         *horizonclient.Client
	Network         string
	StartingBalance string
	BaseFee         int64
	Dispatcher      MinionDispatcher

	// Uninitialized.
	forceRefreshSequence bool
}

// NewMinion ...
func NewMinion(
	a Account,
	kp *keypair.Full,
	botAccount txnbuild.Account,
	botKeypair *keypair.Full,
	hclient *horizonclient.Client,
	network string,
	startingBalance string,
	baseFee int64,
) Minion {
	d := &BasicMinionDispatcher{}
	return Minion{
		a,
		kp,
		botAccount,
		botKeypair,
		hclient,
		network,
		startingBalance,
		baseFee,
		d,
		false,
	}
}

// Run reads a payment destination address and an output channel. It attempts
// to pay that address and submits the result to the channel.
func (minion *Minion) Run(destAddress string, resultChan chan SubmitResult) {
	err := minion.Dispatcher.CheckSequenceRefresh(minion, minion.Horizon)
	if err != nil {
		resultChan <- SubmitResult{
			maybeTransactionSuccess: nil,
			maybeErr:                errors.Wrap(err, "checking minion seq"),
		}
		return
	}
	txStr, err := minion.makeTx(destAddress)
	if err != nil {
		resultChan <- SubmitResult{
			maybeTransactionSuccess: nil,
			maybeErr:                errors.Wrap(err, "making payment tx"),
		}
		return
	}
	succ, err := minion.Dispatcher.SubmitTransaction(minion, minion.Horizon, txStr)
	resultChan <- SubmitResult{
		maybeTransactionSuccess: succ,
		maybeErr:                errors.Wrap(err, "submitting tx to minion"),
	}
}

func (minion *Minion) checkHandleBadSequence(err *horizonclient.Error) {
	resCode, e := err.ResultCodes()
	isTxBadSeqCode := e == nil && resCode.TransactionCode == "tx_bad_seq"
	if !isTxBadSeqCode {
		return
	}
	minion.forceRefreshSequence = true
}

func (minion *Minion) makeTx(destAddress string) (string, error) {
	createAccountOp := txnbuild.CreateAccount{
		Destination:   destAddress,
		SourceAccount: minion.BotAccount,
		Amount:        minion.StartingBalance,
	}
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        minion.Account,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{&createAccountOp},
			BaseFee:              minion.BaseFee,
			Timebounds:           txnbuild.NewInfiniteTimeout(),
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "unable to build tx")
	}

	tx, err = tx.Sign(minion.Network, minion.Keypair, minion.BotKeypair)
	if err != nil {
		return "", errors.Wrap(err, "unable to sign tx")
	}

	txe, err := tx.Base64()
	if err != nil {
		return "", errors.Wrap(err, "unable to serialize")
	}

	// Increment the in-memory sequence number, since the tx will be submitted.
	_, err = minion.Account.IncrementSequenceNumber()
	if err != nil {
		return "", errors.Wrap(err, "incrementing minion seq")
	}
	return txe, err
}
