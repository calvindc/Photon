package blockchain

import (
	"strings"

	"errors"
	"math/big"

	"bytes"
	"fmt"

	"encoding/hex"

	"github.com/SmartMeshFoundation/SmartRaiden/log"
	"github.com/SmartMeshFoundation/SmartRaiden/network/rpc/contracts"
	"github.com/SmartMeshFoundation/SmartRaiden/params"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

/*
query and subscribe may get other contract's event which has the same name and arguments.
*/
var eventTokenAddedID common.Hash
var eventChannelNewID common.Hash
var eventChannelDeletedID common.Hash
var eventChannelNewBalanceID common.Hash
var eventChannelClosedID common.Hash
var eventTransferUpdatedID common.Hash
var eventChannelSettledID common.Hash
var eventChannelSecretRevealedID common.Hash
var eventAddressRegisteredID common.Hash
var errEventNotMatch = errors.New("")

//Event is common part of a blockchain event
type chainEvent struct {
	EventName       string
	BlockNumber     int64
	TxIndex         uint
	TxHash          common.Hash
	ContractAddress common.Address
}

func (ce *chainEvent) Name() string {
	return ce.EventName
}
func initEventWithLog(el *types.Log, e *chainEvent) {
	e.BlockNumber = int64(el.BlockNumber)
	e.TxIndex = el.TxIndex
	e.TxHash = el.TxHash
	e.ContractAddress = el.Address
}

//EventTokenAdded is a new token event
type EventTokenAdded struct {
	chainEvent
	TokenAddress          common.Address
	ChannelManagerAddress common.Address
}

func debugPrintLog(l *types.Log) {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "{\nblocknumber=%d,txIndex=%d,Index=%d,Address=%s\n",
		l.BlockNumber, l.TxIndex, l.Index, utils.APex(l.Address))
	for i, t := range l.Topics {
		fmt.Fprintf(w, "topics[%d]=%s\n", i, t.String())
	}
	fmt.Fprintf(w, "data:\n%s\n}", hex.Dump(l.Data))
	log.Trace(string(w.Bytes()))
}
func newEventTokenAdded(el *types.Log) (e *EventTokenAdded, err error) {
	if len(el.Data) < 64 {
		err = errEventNotMatch
		return
	}
	e = &EventTokenAdded{}
	e.EventName = params.NameTokenAdded
	if eventTokenAddedID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.RegistryABI))
		if err != nil {
			return nil, err
		}
		eventTokenAddedID = parsed.Events[e.EventName].Id()
	}
	if eventTokenAddedID != el.Topics[0] {
		log.Crit("newEventTokenAdded with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.TokenAddress = common.BytesToAddress(el.Data[12:32])          // the first 32byte is tokenaddress
	e.ChannelManagerAddress = common.BytesToAddress(el.Data[44:64]) //the second 32byte is channelManagerAddress
	return
}

//EventChannelNew event on contract
type EventChannelNew struct {
	chainEvent
	NettingChannelAddress common.Address
	Participant1          common.Address
	Participant2          common.Address
	SettleTimeout         int
}

func newEventEventChannelNew(el *types.Log) (e *EventChannelNew, err error) {
	if len(el.Data) < 128 {
		err = errEventNotMatch
		return
	}
	e = &EventChannelNew{}
	e.EventName = params.NameChannelNew
	if eventChannelNewID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.ChannelManagerContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelNewID = parsed.Events[e.EventName].Id()
	}
	if eventChannelNewID != el.Topics[0] {
		log.Crit("newEventEventChannelNew with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.NettingChannelAddress = common.BytesToAddress(el.Data[12:32]) //the first 32byte is tokenaddress
	e.Participant1 = common.BytesToAddress(el.Data[44:64])          //the second 32byte is channelManagerAddress
	e.Participant2 = common.BytesToAddress(el.Data[76:96])
	t := new(big.Int)
	t.SetBytes(el.Data[96:128])
	e.SettleTimeout = int(t.Int64())
	return
}

//EventChannelDeleted on contract
type EventChannelDeleted struct {
	chainEvent
	CallerAddress common.Address
	Partener      common.Address
}

func newEventChannelDeleted(el *types.Log) (e *EventChannelDeleted, err error) {
	if len(el.Data) < 64 {
		err = errEventNotMatch
		return
	}
	e = &EventChannelDeleted{}
	e.EventName = params.NameChannelDeleted
	if eventChannelDeletedID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.ChannelManagerContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelDeletedID = parsed.Events[e.EventName].Id()
	}
	if eventChannelDeletedID != el.Topics[0] {
		log.Crit("newEventEventChannelNew with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.CallerAddress = common.BytesToAddress(el.Data[12:32]) //the first 32byte is tokenaddress
	e.Partener = common.BytesToAddress(el.Data[44:64])      //the second 32byte is channelManagerAddress
	return
}

//EventChannelNewBalance event on contract
type EventChannelNewBalance struct {
	chainEvent
	TokenAddress       common.Address
	ParticipantAddress common.Address
	Balance            *big.Int
}

func newEventChannelNewBalance(el *types.Log) (e *EventChannelNewBalance, err error) {
	if len(el.Data) < 96 {
		err = errEventNotMatch
		return
	}
	e = &EventChannelNewBalance{}
	e.EventName = params.NameChannelNewBalance
	if eventChannelNewBalanceID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.NettingChannelContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelNewBalanceID = parsed.Events[e.EventName].Id()
	}
	if eventChannelNewBalanceID != el.Topics[0] {
		log.Crit("newEventChannelNewBalance with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.TokenAddress = common.BytesToAddress(el.Data[12:32])       //the first 32byte is tokenaddress
	e.ParticipantAddress = common.BytesToAddress(el.Data[44:64]) //the second 32byte is channelManagerAddress
	t := new(big.Int)
	t.SetBytes(el.Data[64:96])
	e.Balance = t
	return
}

//EventChannelClosed event on contract
//event ChannelClosed(address closing_address, uint block_number);
type EventChannelClosed struct {
	chainEvent
	ClosingAddress common.Address
}

func newEventChannelClosed(el *types.Log) (e *EventChannelClosed, err error) {
	if len(el.Data) < 32 {
		err = errEventNotMatch
		return
	}
	e = &EventChannelClosed{}
	e.EventName = params.NameChannelClosed
	if eventChannelClosedID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.NettingChannelContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelClosedID = parsed.Events[e.EventName].Id()
	}
	if eventChannelClosedID != el.Topics[0] {
		log.Crit("newEventChannelClosed with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.ClosingAddress = common.BytesToAddress(el.Data[12:32]) //the first 32byte is tokenaddress
	return
}

//EventTransferUpdated contract event
//event TransferUpdated(address node_address, uint block_number);
type EventTransferUpdated struct {
	chainEvent
	NodeAddress common.Address
}

func newEventTransferUpdated(el *types.Log) (e *EventTransferUpdated, err error) {
	if len(el.Data) < 32 {
		err = errEventNotMatch
		return
	}
	e = &EventTransferUpdated{}
	e.EventName = params.NameTransferUpdated
	if eventTransferUpdatedID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.NettingChannelContractABI))
		if err != nil {
			return nil, err
		}
		eventTransferUpdatedID = parsed.Events[e.EventName].Id()
	}
	if eventTransferUpdatedID != el.Topics[0] {
		log.Crit("NewEventTransferUpdatedd with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.NodeAddress = common.BytesToAddress(el.Data[12:32]) //the first 32byte is tokenaddress
	return
}

//EventChannelSettled contract event
//event ChannelSettled(uint block_number);
type EventChannelSettled struct {
	chainEvent
}

func newEventChannelSettled(el *types.Log) (e *EventChannelSettled, err error) {
	e = &EventChannelSettled{}
	e.EventName = params.NameChannelSettled
	if eventChannelSettledID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.NettingChannelContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelSettledID = parsed.Events[e.EventName].Id()
	}
	if eventChannelSettledID != el.Topics[0] {
		log.Crit("NewEventChannelSettledd with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	return
}

//EventChannelSecretRevealed event contract
//event ChannelSecretRevealed(bytes32 secret, address receiver_address);
type EventChannelSecretRevealed struct {
	chainEvent
	Secret          common.Hash
	ReceiverAddress common.Address
}

func newEventChannelSecretRevealed(el *types.Log) (e *EventChannelSecretRevealed, err error) {
	if len(el.Data) < 64 {
		err = errEventNotMatch
		return
	}
	e = &EventChannelSecretRevealed{}
	e.EventName = params.NameChannelSecretRevealed
	if eventChannelSecretRevealedID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		parsed, err := abi.JSON(strings.NewReader(contracts.NettingChannelContractABI))
		if err != nil {
			return nil, err
		}
		eventChannelSecretRevealedID = parsed.Events[e.EventName].Id()
	}
	if eventChannelSecretRevealedID != el.Topics[0] {
		log.Crit("NewEventChannelSecretRevealedd with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.Secret = common.BytesToHash(el.Data[:32]) //the first 32byte is secret,the second is address
	e.ReceiverAddress = common.BytesToAddress(el.Data[44:64])
	return
}

//EventAddressRegistered contract event
// event AddressRegistered(address indexed eth_address, string socket);
type EventAddressRegistered struct {
	chainEvent
	EthAddress common.Address
	Socket     string
}

//NewEventAddressRegistered new token event
func NewEventAddressRegistered(el *types.Log) (e *EventAddressRegistered, err error) {
	if len(el.Data) < 64 {
		err = errEventNotMatch
		return
	}
	e = &EventAddressRegistered{}
	e.EventName = params.NameAddressRegistered
	if eventAddressRegisteredID == utils.EmptyHash {
		//no error test,the abi is generated by abigen
		var parsed abi.ABI
		parsed, err = abi.JSON(strings.NewReader(contracts.EndpointRegistryABI))
		if err != nil {
			return nil, err
		}
		eventAddressRegisteredID = parsed.Events[e.EventName].Id()
	}
	if eventAddressRegisteredID != el.Topics[0] {
		log.Crit("NewEventAddressRegisteredd with unknown log: ", el)
	}
	initEventWithLog(el, &e.chainEvent)
	e.EthAddress = common.BytesToAddress(el.Topics[1][12:32]) //
	/* Data todo why is  first 32bytes empty?
		Data: ([]uint8) (len=96 cap=96) {
	            00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	            00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 20  |............... |
	            00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	            00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 12  |................|
	            00000040  31 37 32 2e 33 31 2e 37  30 2e 32 38 3a 34 30 30  |172.31.70.28:400|
	            00000050  30 31 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |01..............|
	        },
	*/
	//the first 32 bytes in the data are empty,what does it mean?
	t := new(big.Int)
	t.SetBytes(el.Data[32:64])
	if len(el.Data) < 64+int(t.Int64()) {
		err = errEventNotMatch
		return
	}
	e.Socket = string(el.Data[64 : 64+int(t.Int64())])
	return
}
