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

package modules

import (
	"github.com/palletone/go-palletone/common"
	"math/big"
	"time"
)

type Contract struct {
	Id       common.Hash
	Name     string
	ConType  string // 合约类型： 系统合约 用户合约
	LangCode string // 代码类别

	Sign        []*Authentifier // 单一签名和多方签名
	Code        []byte          // 合约代码。
	CodeHash    common.Hash
	CodeAddress common.Address
	Input       []byte

	CallerAddress common.Address
	caller        common.Address
	self          common.Address // 合約地址
	jumpdests     map[common.Hash][]byte

	value *big.Int

	Args []byte

	status int // 合约状态
	// creator
	creation time.Time // creation  date
}

func (c *Contract) Caller() common.Address { return c.caller }

func (c *Contract) Self() common.Address { return c.self }

func (c *Contract) Jumpdests() map[common.Hash][]byte { return c.jumpdests }

func (c *Contract) Value() *big.Int { return c.value }

func (c *Contract) Status() int { return c.status }
