// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// This file contains some shares testing functionality, common to  multiple
// different files and modules being tested.

package ptn

import (
	"crypto/ecdsa"
	"crypto/rand"
	//"math/big"
	"sync"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/consensus"

	//"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"

	//"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag"
	"fmt"
	"time"
	common2 "github.com/palletone/go-palletone/dag/common"
)

var (
	testBankKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testBank       = crypto.PubkeyToAddress(testBankKey.PublicKey)
)

// newTestProtocolManager creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events.
func newTestProtocolManager(mode downloader.SyncMode, blocks int, newtx chan<- []*modules.Transaction) (*ProtocolManager, ptndb.Database, error) {
	//var dag1 dag.Dag
	//if blocks == 0 {
	//	dag1 = *dag.NewDag()
	//}else{
	//	dag1 = *makedag(blocks)
	//}
	//PrintDag()
	engine := new(consensus.DPOSEngine)
	dag := dag.NewDag()
	typemux := new(event.TypeMux)
	//DbPath := "./data1/leveldb"
	db, _ := ptndb.NewMemDatabase()
	producer := new(mediatorplugin.MediatorPlugin)

	//want (downloader.SyncMode, uint64, txPool, core.ConsensusEngine, *modules.Dag, *event.TypeMux, *ptndb.LDBDatabase)
	pm, err := NewProtocolManager(mode, DefaultConfig.NetworkId, &testTxPool{added: newtx},
		engine, dag, typemux, db, producer)
	if err != nil {
		return nil, nil, err
	}
	pm.Start(1000)
	return pm, db, nil
}

// newTestProtocolManagerMust creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events. In case of an error, the constructor force-
// fails the test.
func newTestProtocolManagerMust(t *testing.T, mode downloader.SyncMode, blocks int, newtx chan<- []*modules.Transaction) (*ProtocolManager, ptndb.Database) {
	pm, db, err := newTestProtocolManager(mode, blocks /*generator,*/, newtx)
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}
	return pm, db
}

func makedag(blocks int)*dag.Dag{
	//PrintDag()
	//fmt.Println("=============")
	dag := dag.NewDag()
	pay := modules.PaymentPayload{
		Inputs:  []modules.Input{},
		Outputs: []modules.Output{},
	}
	holder := common.Address{}
	holder.SetString("P1MEh8GcaAwS3TYTomL1hwcbuhnQDStTmgc")

	msg0 := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay,
	}
	tx := &modules.Transaction{
		TxMessages:   []modules.Message{msg0},
	}
	genesis,_ := common2.NewGenesisUnit(modules.Transactions{tx},time.Now().Unix())
	//fmt.Printf("-----------%+v\n",genesis)
	//fmt.Printf("-------1----%+v\n",genesis.Header().Number.Index)
	authentifiier := &modules.Authentifier{holder.String(),nil,nil,nil}
	genesis.UnitHeader.Authors = authentifiier
	//fmt.Println("--",authentifiier)
	//fmt.Println("--",genesis.UnitHeader.Authors)
	genesis.Header().Number.Index = 0
	dag.SaveDag(*genesis)
	//fmt.Printf("===========111=%#v\n",dag.CurrentHeader())
	//fmt.Printf("===========111=%#v\n",dag.CurrentUnit())
	//fmt.Printf("===========111=%#v\n",dag.GetUnitByNumber(0).Header())
	//fmt.Printf("===========111=%#v\n",dag.CurrentHeader())
	return dag
}

func PrintDag() {
	dag := dag.NewDag()
	chainindex := modules.ChainIndex{
		modules.IDType16{},
		true,
		1,
	}
	header := dag.GetHeaderByNumber(chainindex)
	fmt.Println("header=1111",header)
	unit := dag.CurrentUnit()
	fmt.Printf("0===========111=%#v\n",unit)
	fmt.Printf("0===========111=%#v\n",unit.UnitHeader)
	i := 0
	if len(unit.UnitHeader.ParentsHash) > 0 {
		unit = dag.GetUnit(unit.ParentHash()[0])
		fmt.Printf("%d===========111=%#v\n",i,unit)
		fmt.Printf("%d===========111=%#v\n",i,unit.UnitHeader)
		i++
	}
}


// testTxPool is a fake, helper transaction pool for testing purposes
type testTxPool struct {
	txFeed event.Feed
	pool   []*modules.Transaction        // Collection of all transactions
	added  chan<- []*modules.Transaction // Notification channel for new transactions

	lock sync.RWMutex // Protects the transaction pool
}

// AddRemotes appends a batch of transactions to the pool, and notifies any
// listeners if the addition channel is non nil
func (p *testTxPool) AddRemotes(txs []*modules.Transaction) []error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, txs...)
	if p.added != nil {
		p.added <- txs
	}
	return make([]error, len(txs))
}


// Pending returns all the transactions known to the pool
func (p *testTxPool) Pending() (map[common.Hash]*modules.TxPoolTransaction, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	batches := make(map[common.Hash]*modules.TxPoolTransaction)
	//for _, tx := range p.pool {
	// from, _ := types.Sender(types.HomesteadSigner{}, tx)
	// batches[from] = append(batches[from], tx)
	//}
	//for _, batch := range batches {
	// sort.Sort(types.TxByNonce(batch))
	//}
	return batches, nil
}

func (p *testTxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription{
	return p.txFeed.Subscribe(ch)
}

// newTestTransaction create a new dummy transaction.
func newTestTransaction(from *ecdsa.PrivateKey, nonce uint64, datasize int) *modules.Transaction {
	msg := modules.Message{
		App: modules.APP_PAYMENT,
		//PayloadHash: common.HexToHash("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
		Payload: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	//tx := modules.NewTransaction(nonce, big.NewInt(0), []byte("abc"))
	tx := modules.NewTransaction(
		[]modules.Message{msg, msg, msg},
		12345,
	)

	return tx
}

// testPeer is a simulated peer to allow testing direct network calls.
type testPeer struct {
	net p2p.MsgReadWriter // Network layer reader/writer to simulate remote messaging
	app *p2p.MsgPipeRW    // Application layer reader/writer to simulate the local side
	*peer
}

// newTestPeer creates a new peer registered at the given protocol manager.
func newTestPeer(name string, version int, pm *ProtocolManager, shake bool) (*testPeer, <-chan error) {
	// Create a message pipe to communicate through
	app, net := p2p.MsgPipe()

	// Generate a random id and create the peer
	var id discover.NodeID
	rand.Read(id[:])

	peer := pm.newPeer(version, p2p.NewPeer(id, name, nil), net)

	// Start the peer on a new thread
	errc := make(chan error, 1)
	go func() {
		select {
		case pm.newPeerCh <- peer:
			errc <- pm.handle(peer)
		case <-pm.quitSync:
			errc <- p2p.DiscQuitting
		}
	}()
	tp := &testPeer{app: app, net: net, peer: peer}
	// Execute any implicitly requested handshakes and return
	if shake {
		var (
			//genesis = pm.dag.GetUnitByNumber(0)
			//head  ,_ = pm.dag.GetHeader(genesis.Header().Hash(),0)
			//td      = pm.dag.GetUnitByNumber(0).Number().Index
		)
		//tp.handshake(nil, 0, common.Hash{}, common.Hash{})
	}
	return tp, errc
}

// handshake simulates a trivial handshake that expects the same state from the
// remote side as we are simulating locally.
func (p *testPeer) handshake(t *testing.T, td uint64, head common.Hash, genesis common.Hash) {
	msg := &statusData{
		ProtocolVersion: uint32(p.version),
		NetworkId:       DefaultConfig.NetworkId,
		TD:              td,
		CurrentBlock:    head,
		GenesisBlock:    genesis,
	}
	if err := p2p.ExpectMsg(p.app, StatusMsg, msg); err != nil {
		t.Fatalf("status recv: %v", err)
	}
	if err := p2p.Send(p.app, StatusMsg, msg); err != nil {
		t.Fatalf("status send: %v", err)
	}
}

// close terminates the local side of the peer, notifying the remote protocol
// manager of termination.
func (p *testPeer) close() {
	p.app.Close()
}
