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

package ptn

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/dag/modules"
	"testing"

	"github.com/palletone/go-palletone/ptn/downloader"
)

// Tests that protocol versions and modes of operations are matched up properly.
func TestProtocolCompatibility(t *testing.T) {
	// Define the compatibility chart
	tests := []struct {
		version    uint
		mode       downloader.SyncMode
		compatible bool
	}{
		{0, downloader.FullSync, true},
		{0, downloader.FullSync, true},
		{1, downloader.FullSync, true},
		{0, downloader.FastSync, false},
		{0, downloader.FastSync, false},
		{1, downloader.FastSync, true},
	}
	// Make sure anything we screw up is restored
	backup := ProtocolVersions
	defer func() { ProtocolVersions = backup }()
	// Try all available compatibility configs and check for errors
	for i, tt := range tests {
		ProtocolVersions = []uint{tt.version}
		pm, _, err := newTestProtocolManager(tt.mode, 0, nil)
		if pm != nil {
			defer pm.Stop()
		}
		if (err == nil && !tt.compatible) || (err != nil && tt.compatible) {
			t.Errorf("test %d: compatibility mismatch: have error %v, want compatibility %v", i, err, tt.compatible)
		}
	}
}

// Tests that block headers can be retrieved from a remote chain based on user queries.
//func TestGetBlockHeaders1(t *testing.T) { testGetBlockHeaders(t, 1) }
func testGetBlockHeaders(t *testing.T, protocol int) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	// Create a "random" unknown hash for testing
	var unknown common.Hash
	for i := range unknown {
		unknown[i] = byte(i)
	}
	// Create a batch of tests for various scenarios
	//limit := uint64(downloader.MaxHeaderFetch)
	tests := []struct {
		query  *getBlockHeadersData // The query to execute for header retrieval
		expect []common.Hash        // The hashes of the block whose headers are expected
	}{
		// Check that non existing headers aren't returned
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: unknown}, Amount: 1},
			[]common.Hash{},
		},
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the headers to expect in the response
		headers := []*modules.Header{}
		for _, hash := range tt.expect {
			//headers = append(headers, pm.blockchain.GetBlockByHash(hash).Header())
			hash = hash
			headers = append(headers, pm.dag.CurrentUnit().UnitHeader)
		}
		// Send the hash request and verify the response
		//p2p.Send(peer.app, 0x00, tt.query)
		//fmt.Println(len(headers))
		if err := p2p.ExpectMsg(peer.app, 0x00, nil); err != nil {
			t.Errorf("test %d: headers mismatch: %v", i, err)
		}
		// If the test used number origins, repeat with hashes as the too
		if tt.query.Origin.Hash == (common.Hash{}) {
			index := modules.ChainIndex{
				IsMain:  true,
				Index:   uint64(0),
			}
			index.AssetID.SetBytes([]byte("test"))
			tt.query.Origin.Hash, tt.query.Origin.Number = common.Hash{}, index
			p2p.Send(peer.app, 0x03, tt.query)
			if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
				t.Errorf("test %d: headers mismatch: %v", i, err)
			}
		}
	}
}

// Tests that block contents can be retrieved from a remote chain based on their hashes.
//func TestGetBlockBodies62(t *testing.T) { testGetBlockBodies(t, 1) }
//func TestGetBlockBodies1(t *testing.T) { testGetBlockBodies(t, 1) }
func testGetBlockBodies(t *testing.T, protocol int) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, 11, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	// Create a batch of tests for various scenarios
	//limit := downloader.MaxBlockFetch
	tests := []struct {
		random    int           // Number of blocks to fetch randomly from the chain
		explicit  []common.Hash // Explicitly requested blocks
		available []bool        // Availability of explicitly requested blocks
		expected  int           // Total number of existing blocks to expect
	}{
		{1, nil, nil, 1},                                                         // A single random block should be retrievable
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the hashes to request, and the response to expect
		hashes, seen := []common.Hash{}, make(map[int64]bool)
		bodies := []*blockBody{}
		for j := 0; j < tt.random; j++ {
			for {
				//num := rand.Int63n(int64(pm.dag.CurrentUnit().UnitHeader.Number.Index))
				if !seen[0] {
					seen[0] = true

					block := pm.dag.GetUnitByNumber(uint64(0))
					hashes = append(hashes, block.Hash())
					if len(bodies) < tt.expected {
						bodies = append(bodies, &blockBody{Transactions: block.Transactions()})
					}
					break
				}
			}
		}
		for j, hash := range tt.explicit {
			hashes = append(hashes, hash)
			if tt.available[j] && len(bodies) < tt.expected {
				block := pm.dag.GetUnitByHash(hash)
				bodies = append(bodies, &blockBody{Transactions: block.Transactions()})
			}
		}
		pay := modules.PaymentPayload{
			Inputs:  []modules.Input{},
			Outputs: []modules.Output{},
		}
		msg0 := modules.Message{
			App:     modules.APP_PAYMENT,
			Payload: pay,
		}
		tx := &modules.Transaction{
			TxMessages:   []modules.Message{msg0},
		}
		bodies = append(bodies, &blockBody{Transactions: []*modules.Transaction{tx}})


		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x00, hashes)
		if err := p2p.ExpectMsg(peer.app, 0x00, nil); err != nil {
			t.Errorf("test %d: bodies mismatch: %v", i, err)
		}
	}
}
/*
// Tests that the node state database can be retrieved based on hashes.
func TestGetNodeData63(t *testing.T) { testGetNodeData(t, 63) }

func testGetNodeData(t *testing.T, protocol int) {
	// Define three accounts to simulate transactions with
	acc1Key, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	acc2Key, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	acc1Addr := crypto.PubkeyToAddress(acc1Key.PublicKey)
	acc2Addr := crypto.PubkeyToAddress(acc2Key.PublicKey)

	signer := types.HomesteadSigner{}
	// Create a chain generator with some simple transactions (blatantly stolen from @fjl/chain_markets_test)
	generator := func(i int, block *core.BlockGen) {
		switch i {
		case 0:
			// In block 1, the test bank sends account #1 some ether.
			tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(10000), configure.TxGas, nil, nil), signer, testBankKey)
			block.AddTx(tx)
		case 1:
			// In block 2, the test bank sends some more ether to account #1.
			// acc1Addr passes it on to account #2.
			tx1, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(1000), configure.TxGas, nil, nil), signer, testBankKey)
			tx2, _ := types.SignTx(types.NewTransaction(block.TxNonce(acc1Addr), acc2Addr, big.NewInt(1000), configure.TxGas, nil, nil), signer, acc1Key)
			block.AddTx(tx1)
			block.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by account #2.
			block.SetCoinbase(acc2Addr)
			block.SetExtra([]byte("yeehaw"))
		case 3:
			// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
			b2 := block.PrevBlock(1).Header()
			b2.Extra = []byte("foo")
			block.AddUncle(b2)
			b3 := block.PrevBlock(2).Header()
			b3.Extra = []byte("foo")
			block.AddUncle(b3)
		}
	}
	// Assemble the test environment
	pm, db := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()

	// Fetch for now the entire chain db
	hashes := []common.Hash{}
	for _, key := range db.Keys() {
		if len(key) == len(common.Hash{}) {
			hashes = append(hashes, common.BytesToHash(key))
		}
	}
	p2p.Send(peer.app, 0x0d, hashes)
	msg, err := peer.app.ReadMsg()
	if err != nil {
		t.Fatalf("failed to read node data response: %v", err)
	}
	if msg.Code != 0x0e {
		t.Fatalf("response packet code mismatch: have %x, want %x", msg.Code, 0x0c)
	}
	var data [][]byte
	if err := msg.Decode(&data); err != nil {
		t.Fatalf("failed to decode response node data: %v", err)
	}
	// Verify that all hashes correspond to the requested data, and reconstruct a state tree
	for i, want := range hashes {
		if hash := crypto.Keccak256Hash(data[i]); hash != want {
			t.Errorf("data hash mismatch: have %x, want %x", hash, want)
		}
	}
	statedb, _ := ptndb.NewMemDatabase()
	for i := 0; i < len(data); i++ {
		statedb.Put(hashes[i].Bytes(), data[i])
	}
	accounts := []common.Address{testBank, acc1Addr, acc2Addr}
	for i := uint64(0); i <= pm.blockchain.CurrentBlock().NumberU64(); i++ {
		trie, _ := state.New(pm.blockchain.GetBlockByNumber(i).Root(), state.NewDatabase(statedb))

		for j, acc := range accounts {
			state, _ := pm.blockchain.State()
			bw := state.GetBalance(acc)
			bh := trie.GetBalance(acc)

			if (bw != nil && bh == nil) || (bw == nil && bh != nil) {
				t.Errorf("test %d, account %d: balance mismatch: have %v, want %v", i, j, bh, bw)
			}
			if bw != nil && bh != nil && bw.Cmp(bw) != 0 {
				t.Errorf("test %d, account %d: balance mismatch: have %v, want %v", i, j, bh, bw)
			}
		}
	}
}
/*
// Tests that the transaction receipts can be retrieved based on hashes.
func TestGetReceipt63(t *testing.T) { testGetReceipt(t, 63) }

func testGetReceipt(t *testing.T, protocol int) {
	// Define three accounts to simulate transactions with
	acc1Key, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	acc2Key, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	acc1Addr := crypto.PubkeyToAddress(acc1Key.PublicKey)
	acc2Addr := crypto.PubkeyToAddress(acc2Key.PublicKey)

	signer := types.HomesteadSigner{}
	// Create a chain generator with some simple transactions (blatantly stolen from @fjl/chain_markets_test)
	generator := func(i int, block *core.BlockGen) {
		switch i {
		case 0:
			// In block 1, the test bank sends account #1 some ether.
			tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(10000), configure.TxGas, nil, nil), signer, testBankKey)
			block.AddTx(tx)
		case 1:
			// In block 2, the test bank sends some more ether to account #1.
			// acc1Addr passes it on to account #2.
			tx1, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(1000), configure.TxGas, nil, nil), signer, testBankKey)
			tx2, _ := types.SignTx(types.NewTransaction(block.TxNonce(acc1Addr), acc2Addr, big.NewInt(1000), configure.TxGas, nil, nil), signer, acc1Key)
			block.AddTx(tx1)
			block.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by account #2.
			block.SetCoinbase(acc2Addr)
			block.SetExtra([]byte("yeehaw"))
		case 3:
			// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
			b2 := block.PrevBlock(1).Header()
			b2.Extra = []byte("foo")
			block.AddUncle(b2)
			b3 := block.PrevBlock(2).Header()
			b3.Extra = []byte("foo")
			block.AddUncle(b3)
		}
	}
	// Assemble the test environment
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()

	// Collect the hashes to request, and the response to expect
	hashes, receipts := []common.Hash{}, []types.Receipts{}
	for i := uint64(0); i <= pm.blockchain.CurrentBlock().NumberU64(); i++ {
		block := pm.blockchain.GetBlockByNumber(i)

		hashes = append(hashes, block.Hash())
		receipts = append(receipts, pm.blockchain.GetReceiptsByHash(block.Hash()))
	}
	// Send the hash request and verify the response
	p2p.Send(peer.app, 0x0f, hashes)
	if err := p2p.ExpectMsg(peer.app, 0x10, receipts); err != nil {
		t.Errorf("receipts mismatch: %v", err)
	}
}


// Tests that post ptn protocol handshake, DAO fork-enabled clients also execute
// a DAO "challenge" verifying each others' DAO fork headers to ensure they're on
// compatible chains.
func TestDAOChallengeNoVsNo(t *testing.T)       { testDAOChallenge(t, false, false, false) }
func TestDAOChallengeNoVsPro(t *testing.T)      { testDAOChallenge(t, false, true, false) }
func TestDAOChallengeProVsNo(t *testing.T)      { testDAOChallenge(t, true, false, false) }
func TestDAOChallengeProVsPro(t *testing.T)     { testDAOChallenge(t, true, true, false) }
func TestDAOChallengeNoVsTimeout(t *testing.T)  { testDAOChallenge(t, false, false, true) }
func TestDAOChallengeProVsTimeout(t *testing.T) { testDAOChallenge(t, true, true, true) }

func testDAOChallenge(t *testing.T, localForked, remoteForked bool, timeout bool) {
	// Reduce the DAO handshake challenge timeout
	if timeout {
		defer func(old time.Duration) { daoChallengeTimeout = old }(daoChallengeTimeout)
		daoChallengeTimeout = 500 * time.Millisecond
	}

	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil)
	//if err != nil {
	//	t.Fatalf("failed to start test protocol manager: %v", err)
	//}
	pm.Start(1000)
	defer pm.Stop()
	// Connect a new peer and check that we receive the DAO challenge
	peer, _ := newTestPeer("peer", 1, pm, true)
	defer peer.close()

	//challenge := &getBlockHeadersData{
	//	Origin:  hashOrNumber{},
	//	Amount:  1,
	//	Skip:    0,
	//	Reverse: true,
	//}
	challenges := make([]getBlockHeadersData,14)
	//challenges[0] = *challenge
	if err := p2p.ExpectMsg(peer.app, StatusMsg, challenges); err != nil {
		t.Fatalf("challenge mismatch: %v", err)
	}
	// Create a block to reply to the challenge if no timeout is simulated
	if !timeout {

		if err := p2p.Send(peer.app, BlockHeadersMsg, []*modules.Header{}); err != nil {
			t.Fatalf("failed to answer challenge: %v", err)
		}
		time.Sleep(100 * time.Millisecond) // Sleep to avoid the verification racing with the drops
	} else {
		// Otherwise wait until the test timeout passes
		time.Sleep(daoChallengeTimeout + 500*time.Millisecond)
	}
	// Verify that depending on fork side, the remote peer is maintained or dropped
	if localForked == remoteForked && !timeout {
		if peers := pm.peers.Len(); peers != 1 {
			t.Fatalf("peer count mismatch: have %d, want %d", peers, 1)
		}
	} else {
		if peers := pm.peers.Len(); peers != 0 {
			t.Fatalf("peer count mismatch: have %d, want %d", peers, 0)
		}
	}
}
*/
