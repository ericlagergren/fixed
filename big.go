package fixed

import (
	"math/big"
	"sync"
	"sync/atomic"
)

// bigMulPow10 sets x = x * 10^n.
func bigMulPow10(x *big.Int, n uint) {
	if x.Sign() != 0 && n != 0 {
		x.Mul(x, bigPow10(n))
	}
}

// bigPowTabLen is the largest cached power for big.Ints.
const bigPowTabLen = 1e5

var (
	bigMu       sync.Mutex // protects writes to bigPow10Tab
	bigPow10Tab atomic.Pointer[[]*big.Int]
)

// bigPow10 computes 10^n.
//
// The result must not be modified.
func bigPow10(n uint) *big.Int {
	tab := *bigPow10Tab.Load()

	if n < uint(len(tab)) {
		return tab[n]
	}

	// Too large for our table.
	if n >= bigPowTabLen {
		// As an optimization, we don't need to start from
		// scratch each time. Start from the largest term we've
		// found so far.
		partial := tab[len(tab)-1]
		p := new(big.Int).SetUint64(uint64(n) - uint64(len(tab)-1))
		return p.Mul(partial, p.Exp(big.NewInt(10), p, nil))
	}
	return growBigTen(n)
}

func growBigTen(n uint) *big.Int {
	// We need to expand our table to contain the value for
	// 10^n.
	bigMu.Lock()
	defer bigMu.Unlock()

	tab := *bigPow10Tab.Load()

	// Look again in case the table was rebuilt before we grabbed
	// the lock.
	if n < uint(len(tab)) {
		return tab[n]
	}
	// n < BigTabLen

	newLen := uint(len(tab) * 2)
	for newLen <= n {
		newLen *= 2
	}
	if newLen > bigPowTabLen {
		newLen = bigPowTabLen
	}
	for i := uint(len(tab)); i < newLen; i++ {
		tab = append(tab, new(big.Int).Mul(tab[i-1], big.NewInt(10)))
	}

	bigPow10Tab.Store(&tab)
	return tab[n]
}

func init() {
	bigPow10Tab.Store(&[]*big.Int{
		new(big.Int).SetUint64(0),
		new(big.Int).SetUint64(1),
		new(big.Int).SetUint64(10),
		new(big.Int).SetUint64(100),
		new(big.Int).SetUint64(1000),
		new(big.Int).SetUint64(10000),
		new(big.Int).SetUint64(100000),
		new(big.Int).SetUint64(1000000),
		new(big.Int).SetUint64(10000000),
		new(big.Int).SetUint64(100000000),
		new(big.Int).SetUint64(1000000000),
		new(big.Int).SetUint64(10000000000),
		new(big.Int).SetUint64(100000000000),
		new(big.Int).SetUint64(1000000000000),
		new(big.Int).SetUint64(10000000000000),
		new(big.Int).SetUint64(100000000000000),
		new(big.Int).SetUint64(1000000000000000),
		new(big.Int).SetUint64(10000000000000000),
		new(big.Int).SetUint64(100000000000000000),
		new(big.Int).SetUint64(1000000000000000000),
		new(big.Int).SetUint64(10000000000000000000),
	})
}
