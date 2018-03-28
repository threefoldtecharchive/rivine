package types

import (
	"math/big"
	"testing"

	"github.com/rivine/rivine/crypto"
)

// TestTargetAdd probes the Add function of the target type.
func TestTargetAdd(t *testing.T) {
	cts := DefaultChainConstants()

	var target3, target5, target10 Target
	target3[crypto.HashSize-1] = 3
	target5[crypto.HashSize-1] = 5
	target10[crypto.HashSize-1] = 10

	expect5 := target10.AddDifficulties(target10, cts.RootDepth)
	if expect5 != target5 {
		t.Error("Target.Add not working as expected")
	}
	expect3 := target10.AddDifficulties(target5, cts.RootDepth)
	if expect3 != target3 {
		t.Error("Target.Add not working as expected")
	}
}

// TestTargetCmp probes the Cmp function of the target type.
func TestTargetCmp(t *testing.T) {
	var target1, target2 Target
	target1[crypto.HashSize-1] = 1
	target2[crypto.HashSize-1] = 2

	if target1.Cmp(target2) != -1 {
		t.Error("Target.Cmp not behaving as expected")
	}
	if target2.Cmp(target2) != 0 {
		t.Error("Target.Cmp not behaving as expected")
	}
	if target2.Cmp(target1) != 1 {
		t.Error("Target.Cmp not behaving as expected")
	}
}

// TestTargetDifficulty probes the Difficulty function of the target type.
// TODO: target.Difficulty() has been changed to return a new Difficutly type: https://github.com/rivine/rivine/commit/6c00f3be7798c017b84434a97f8b1cd895bf734a#diff-3e315e2069bf2cb7e545291ccc110377
// Need to rework test to compile and work with the new type
func TestTargetDifficulty(t *testing.T) {
	cts := DefaultChainConstants()

	var target1, target2, target3 Target
	target2[crypto.HashSize-1] = 1
	target3[crypto.HashSize-1] = 2

	expDifficulty1 := NewDifficulty(cts.RootDepth.Int())
	expDifficulty2 := NewDifficulty(cts.RootDepth.Int())
	expDifficulty3 := NewDifficulty(cts.RootDepth.Int()).Div64(2)

	if difficulty := target1.Difficulty(cts.RootDepth); difficulty.Cmp(expDifficulty1) != 0 {
		t.Errorf("Expected difficulty %v, got %v", expDifficulty1, difficulty)
	}
	if difficulty := target2.Difficulty(cts.RootDepth); difficulty.Cmp(expDifficulty2) != 0 {
		t.Errorf("Expected difficulty %v, got %v", expDifficulty2, difficulty)
	}
	if difficulty := target3.Difficulty(cts.RootDepth); difficulty.Cmp(expDifficulty3) != 0 {
		t.Errorf("Expected difficulty %v, got %v", expDifficulty3, difficulty)
	}
}

// TestTargetInt probes the Int function of the target type.
func TestTargetInt(t *testing.T) {
	var target Target
	target[crypto.HashSize-1] = 1

	b := target.Int()
	if b.Cmp(big.NewInt(1)) != 0 {
		t.Error("Target.Int did not work correctly")
	}
}

// TestTargetIntToTarget probes the IntToTarget function of the target type.
func TestTargetIntToTarget(t *testing.T) {
	cts := DefaultChainConstants()

	var target Target
	target[crypto.HashSize-1] = 5
	b := big.NewInt(5)
	if IntToTarget(b, cts.RootDepth) != target {
		t.Error("IntToTarget not working as expected")
	}
}

// TestTargetInverse probes the Inverse function of the target type.
func TestTargetInverse(t *testing.T) {
	var target Target
	target[crypto.HashSize-1] = 2

	r := target.Inverse()
	if r.Num().Cmp(big.NewInt(1)) != 0 {
		t.Error("Target.Rat did not work as expected")
	}
	if r.Denom().Cmp(big.NewInt(2)) != 0 {
		t.Error("Target.Rat did not work as expected")
	}
}

// TestTargetMul probes the Mul function of the target type.
func TestTargetMul(t *testing.T) {
	cts := DefaultChainConstants()

	var target2, target6, target10, target14, target20 Target
	target2[crypto.HashSize-1] = 2
	target6[crypto.HashSize-1] = 6
	target10[crypto.HashSize-1] = 10
	target14[crypto.HashSize-1] = 14
	target20[crypto.HashSize-1] = 20

	// Multiplying the difficulty of a target at '10' by 5 will yield a target
	// of '2'. Similar math follows for the remaining checks.
	expect2 := target10.MulDifficulty(big.NewRat(5, 1), cts.RootDepth)
	if expect2 != target2 {
		t.Error(expect2)
		t.Error(target2)
		t.Error("Target.Mul did not work as expected")
	}
	expect6 := target10.MulDifficulty(big.NewRat(3, 2), cts.RootDepth)
	if expect6 != target6 {
		t.Error("Target.Mul did not work as expected")
	}
	expect14 := target10.MulDifficulty(big.NewRat(7, 10), cts.RootDepth)
	if expect14 != target14 {
		t.Error("Target.Mul did not work as expected")
	}
	expect20 := target10.MulDifficulty(big.NewRat(1, 2), cts.RootDepth)
	if expect20 != target20 {
		t.Error("Target.Mul did not work as expected")
	}
}

// TestTargetRat probes the Rat function of the target type.
func TestTargetRat(t *testing.T) {
	var target Target
	target[crypto.HashSize-1] = 3

	r := target.Rat()
	if r.Num().Cmp(big.NewInt(3)) != 0 {
		t.Error("Target.Rat did not work as expected")
	}
	if r.Denom().Cmp(big.NewInt(1)) != 0 {
		t.Error("Target.Rat did not work as expected")
	}
}

// TestTargetOverflow checks that IntToTarget will return a maximum target if
// there is an overflow.
func TestTargetOverflow(t *testing.T) {
	cts := DefaultChainConstants()

	largeInt := new(big.Int).Lsh(big.NewInt(1), 260)
	expectRoot := IntToTarget(largeInt, cts.RootDepth)
	if expectRoot != cts.RootDepth {
		t.Error("IntToTarget does not properly handle overflows")
	}
}

// TestTargetNegativeIntToTarget tries to create a negative target using
// IntToTarget.
func TestTargetNegativeIntToTarget(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cts := DefaultChainConstants()

	// In debug mode, attempting to create a negative target should trigger a
	// panic.
	defer func() {
		r := recover()
		if r != ErrNegativeTarget {
			t.Error("no panic occurred when trying to create a negative target")
		}
	}()
	b := big.NewInt(-3)
	_ = IntToTarget(b, cts.RootDepth)
}

// TestTargetNegativeRatToTarget tries to create a negative target using
// RatToTarget.
func TestTargetNegativeRatToTarget(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// In debug mode, attempting to create a negative target should trigger a
	// panic.
	defer func() {
		r := recover()
		if r != ErrNegativeTarget {
			t.Error("no panic occurred when trying to create a negative target")
		}
	}()
	r := big.NewRat(3, -5)
	cts := DefaultChainConstants()
	_ = RatToTarget(r, cts.RootDepth)
}
