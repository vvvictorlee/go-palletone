/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package adaptor

type Config struct {
	Btc BTC
	Eth ETH
}

type BTC struct {
	NetID     int
	Host      string
	RPCUser   string
	RPCPasswd string
	CertPath  string

	ChaincodeKeys map[string]string
	AddressKeys   map[string]string
}
type ETH struct {
	NetID  int
	Rawurl string

	ChaincodeKeys map[string]string
	AddressKeys   map[string]string
}

var DefaultConfig = Config{
	Btc: BTC{
		NetID:     1,
		Host:      "localhost:18332",
		RPCUser:   "zxl",
		RPCPasswd: "123456",
		ChaincodeKeys: map[string]string{
			"SampleSysCC": "123",
		},
		AddressKeys: map[string]string{
			"P1XXX": "123",
		},
	},
	Eth: ETH{
		NetID:  1,
		Rawurl: "\\\\.\\pipe\\geth.ipc",
		ChaincodeKeys: map[string]string{
			"SampleSysCC": "123",
		},
		AddressKeys: map[string]string{
			"P1XXX": "123",
		},
	},
}
