package stake

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

// transfer coins
type transferFn func(from sdk.Actor, to sdk.Actor, coins coin.Coins) abci.Result

// default transfer runs full DeliverTX
func defaultTransferFn(ctx sdk.Context, store state.SimpleDB, dispatch sdk.Deliver) transferFn {
	return func(sender, receiver sdk.Actor, coins coin.Coins) (res abci.Result) {
		// Move coins from the delegator account to the candidate lock account
		send := coin.NewSendOneTx(sender, receiver, coins)

		// If the deduction fails (too high), abort the command
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return abci.ErrInsufficientFunds.AppendLog(err.Error())
		}
		return
	}
}

const (
	candidateBondKey = iota
	delegatorBondKey
	paramKey
)

// LoadCandidateBonds - loads the candidate bond set
// TODO ultimately this function should be made unexported... being used right now
// for patchwork of tick functionality therefor much easier if exported until
// the new SDK is created
func LoadCandidateBonds(store state.SimpleDB) (candidateBonds CandidateBonds) {
	b := store.Get([]byte{candidateBondKey})
	if b == nil {
		return
	}
	err := wire.ReadBinaryBytes(b, &candidateBonds)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return
}

func saveCandidateBonds(store state.SimpleDB, candidateBonds CandidateBonds) {
	b := wire.BinaryBytes(candidateBonds)
	store.Set([]byte{candidateBondKey}, b)
}

/////////////////////////////////////////////////////////////////////////////////

func loadDelegatorBondsKey(delegator sdk.Actor) []byte {
	delegatorBytes := wire.BinaryBytes(&delegator)
	return append([]byte{delegatorBondKey}, delegatorBytes...)
}
func getDelegatorFromKey(key []byte) (delegator sdk.Actor, err error) {
	err = wire.ReadBinaryBytes(key[1:], &delegator)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func saveDelegatorBonds(store state.SimpleDB, delegator sdk.Actor, bonds DelegatorBonds) {
	bondsBytes := wire.BinaryBytes(bonds)
	store.Set(loadDelegatorBondsKey(delegator), bondsBytes)
}

func loadDelegatorBonds(store state.SimpleDB,
	delegator sdk.Actor) (bonds DelegatorBonds, err error) {

	delegatorBytes := store.Get(loadDelegatorBondsKey(delegator))
	if delegatorBytes == nil {
		return
	}
	return readDelegatorBonds(delegatorBytes)
}

func readDelegatorBonds(delegatorBytes []byte) (bonds DelegatorBonds, err error) {
	err = wire.ReadBinaryBytes(delegatorBytes, &bonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func removeDelegatorBonds(store state.SimpleDB, delegator sdk.Actor) {
	store.Remove(loadDelegatorBondsKey(delegator))
}

/////////////////////////////////////////////////////////////////////////////////

// load/save the global staking params
func loadParams(store state.SimpleDB) (params Params) {
	b := store.Get([]byte{paramKey})
	if b == nil {
		return defaultParams()
	}
	err := wire.ReadBinaryBytes(b, &params)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return
}
func saveParams(store state.SimpleDB, params Params) {
	b := wire.BinaryBytes(params)
	store.Set([]byte{paramKey}, b)
}
