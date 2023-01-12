package fixed

import (
	"math/big"
	"testing"

	"github.com/ericlagergren/testutil"
	"golang.org/x/exp/rand"
)

func bigMask(n uint) *big.Int {
	v := new(big.Int).Lsh(big.NewInt(1), n) // 1<<n
	return v.Sub(v, big.NewInt(1))          // 1<<n - 1
}

func randBool() bool {
	return rand.Intn(2) == 0
}

func randUint64() uint64 {
	return rand.Uint64()
}

var sink struct {
	string   string
	uint64   uint64
	Uint96   Uint96
	Uint128  Uint128
	Uint192  Uint192
	Uint256  Uint256
	Uint512  Uint512
	Uint1024 Uint1024
	Uint2048 Uint2048
}

func TestInlining(t *testing.T) {
	testutil.TestInlining(t, "github.com/ericlagergren/fixed",
		"ParseUint1024",
		"ParseUint128",
		"ParseUint192",
		"ParseUint2048",
		"ParseUint256",
		"ParseUint512",
		"ParseUint96",
		"U1024",
		"U128",
		"U192",
		"U2048",
		"U256",
		"U512",
		"U96",
		"Uint1024.Equal",
		"Uint1024.IsZero",
		"Uint1024.Size",
		"Uint128.Add",
		"Uint128.AddCheck",
		"Uint128.And",
		"Uint128.BitLen",
		"Uint128.Bytes",
		"Uint128.Cmp",
		"Uint128.Equal",
		"Uint128.GoString",
		"Uint128.IsZero",
		"Uint128.LeadingZeros",
		"Uint128.Lsh",
		"Uint128.Mul",
		"Uint128.Or",
		"Uint128.Rsh",
		"Uint128.Size",
		"Uint128.Sub",
		"Uint128.SubCheck",
		"Uint128.Xor",
		"Uint128.add64",
		"Uint128.addCheck64",
		"Uint128.cmp64",
		"Uint128.max",
		"Uint128.mul64",
		"Uint128.mulCheck64",
		"Uint128.sub64",
		"Uint128.uint192",
		"Uint192.Add",
		"Uint192.AddCheck",
		"Uint192.And",
		"Uint192.BitLen",
		"Uint192.Cmp",
		"Uint192.Equal",
		"Uint192.GoString",
		"Uint192.IsZero",
		"Uint192.LeadingZeros",
		"Uint192.Or",
		"Uint192.Size",
		"Uint192.Sub",
		"Uint192.SubCheck",
		"Uint192.Xor",
		"Uint192.add64",
		"Uint192.addCheck64",
		"Uint192.cmp64",
		"Uint192.hi128",
		"Uint192.low128",
		"Uint192.max",
		"Uint192.mul64",
		"Uint2048.Equal",
		"Uint2048.IsZero",
		"Uint2048.LeadingZeros",
		"Uint2048.Size",
		"Uint256.Add",
		"Uint256.AddCheck",
		"Uint256.And",
		"Uint256.BitLen",
		"Uint256.Bytes",
		"Uint256.Equal",
		"Uint256.GoString",
		"Uint256.IsZero",
		"Uint256.LeadingZeros",
		"Uint256.Or",
		"Uint256.Size",
		"Uint256.Sub",
		"Uint256.SubCheck",
		"Uint256.Xor",
		"Uint256.add64",
		"Uint256.addCheck64",
		"Uint256.cmp64",
		"Uint256.max",
		"Uint256.sub64",
		"Uint512.And",
		"Uint512.Bytes",
		"Uint512.Equal",
		"Uint512.IsZero",
		"Uint512.LeadingZeros",
		"Uint512.Or",
		"Uint512.Size",
		"Uint512.Xor",
		"Uint96.Add",
		"Uint96.AddCheck",
		"Uint96.And",
		"Uint96.BitLen",
		"Uint96.Cmp",
		"Uint96.Equal",
		"Uint96.GoString",
		"Uint96.IsZero",
		"Uint96.LeadingZeros",
		"Uint96.Lsh",
		"Uint96.Mul",
		"Uint96.Or",
		"Uint96.Rsh",
		"Uint96.Size",
		"Uint96.String",
		"Uint96.Sub",
		"Uint96.SubCheck",
		"Uint96.Xor",
		"Uint96.add64",
		"Uint96.addCheck64",
		"Uint96.cmp64",
		"Uint96.max",
		"Uint96.mul64",
		"Uint96.mulCheck64",
		"cloneString",
		"digits",
		"lower",
		"mulAddWWW",
		"mulAddWWWW",
		"rangeError",
		"syntaxError",
		"u256",
		"Uint1024.LeadingZeros",
	)
}
