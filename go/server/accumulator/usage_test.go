package accumulator

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CommonUse(t *testing.T) {
	curve := curves.BLS12381(&curves.PointBls12381G1{})
	sk, _ := new(SecretKey).New(curve, []byte("1234567890"))
	pk, _ := sk.GetPublicKey(curve)

	illigal := curve.Scalar.Hash([]byte("illigal"))

	prog1 := curve.Scalar.Hash([]byte("prog1"))
	prog2 := curve.Scalar.Hash([]byte("prog2"))
	prog3 := curve.Scalar.Hash([]byte("prog3"))

	//progs := []Element{prog1, prog2, prog3}

	team1 := curve.Scalar.Hash([]byte("team1"))
	team2 := curve.Scalar.Hash([]byte("team2"))
	//teams := []Element{team1, team2}

	cto := curve.Scalar.Hash([]byte("cto"))
	allElems := []Element{prog1, prog2, prog3, team1, team2, cto}

	allAcc, err := new(Accumulator).New(curve)
	require.NoError(t, err)
	allAcc, err = allAcc.AddElements(sk, allElems)
	require.NoError(t, err)

	witProg1, err := new(MembershipWitness).New(prog1, allAcc, sk)
	witProg2, err := new(MembershipWitness).New(prog2, allAcc, sk)
	witProg3, err := new(MembershipWitness).New(prog3, allAcc, sk)

	witTeam2, err := new(MembershipWitness).New(team2, allAcc, sk)
	witTeam1, err := new(MembershipWitness).New(team1, allAcc, sk)

	witCTO, err := new(MembershipWitness).New(cto, allAcc, sk)

	witIlligal, err := new(MembershipWitness).New(illigal, allAcc, sk)

	allWits := []*MembershipWitness{witProg1, witProg2, witProg3, witTeam1, witTeam2, witCTO}

	for i := range allElems {
		err = allWits[i].Verify(pk, allAcc)
		require.NoError(t, err)
	}

	require.Error(t, witIlligal.Verify(pk, allAcc))
}
