// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package groth16

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/hash_to_field"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/pedersen"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/logger"
)

var (
	errPairingCheckFailed         = errors.New("pairing doesn't match")
	errCorrectSubgroupCheckFailed = errors.New("points in the proof are not in the correct subgroup")
)

// Verify verifies a proof with given VerifyingKey and publicWitness
func Verify(proof *Proof, vk *VerifyingKey, publicWitness fr.Vector, opts ...backend.VerifierOption) error {
	opt, err := backend.NewVerifierConfig(opts...)
	if err != nil {
		return fmt.Errorf("new verifier config: %w", err)
	}
	if opt.HashToFieldFn == nil {
		opt.HashToFieldFn = hash_to_field.New([]byte(constraint.CommitmentDst))
	}
	fmt.Println(len(vk.G1.K))
	nbPublicVars := len(vk.G1.K) - len(vk.PublicAndCommitmentCommitted)

	if len(publicWitness) != nbPublicVars-1 {
		return fmt.Errorf("invalid witness size, got %d, expected %d (public - ONE_WIRE)", len(publicWitness), len(vk.G1.K)-1)
	}
	log := logger.Logger().With().Str("curve", vk.CurveID().String()).Str("backend", "groth16").Logger()
	start := time.Now()

	// check that the points in the proof are in the correct subgroup
	if !proof.isValid() {
		return errCorrectSubgroupCheckFailed
	}

	var doubleML curve.GT
	chDone := make(chan error, 1)

	// compute (eKrsδ, eArBs)
	go func() {
		var errML error
		doubleML, errML = curve.MillerLoop([]curve.G1Affine{proof.Krs, proof.Ar}, []curve.G2Affine{vk.G2.deltaNeg, proof.Bs})
		chDone <- errML
		close(chDone)
	}()

	maxNbPublicCommitted := 0
	for _, s := range vk.PublicAndCommitmentCommitted { // iterate over commitments
		maxNbPublicCommitted = utils.Max(maxNbPublicCommitted, len(s))
	}
	commitmentsSerialized := make([]byte, len(vk.PublicAndCommitmentCommitted)*fr.Bytes)
	commitmentPrehashSerialized := make([]byte, curve.SizeOfG1AffineUncompressed+maxNbPublicCommitted*fr.Bytes)
	for i := range vk.PublicAndCommitmentCommitted { // solveCommitmentWire
		copy(commitmentPrehashSerialized, proof.Commitments[i].Marshal())
		offset := curve.SizeOfG1AffineUncompressed
		for j := range vk.PublicAndCommitmentCommitted[i] {
			copy(commitmentPrehashSerialized[offset:], publicWitness[vk.PublicAndCommitmentCommitted[i][j]-1].Marshal())
			offset += fr.Bytes
		}
		opt.HashToFieldFn.Write(commitmentPrehashSerialized[:offset])
		hashBts := opt.HashToFieldFn.Sum(nil)
		opt.HashToFieldFn.Reset()
		nbBuf := fr.Bytes
		if opt.HashToFieldFn.Size() < fr.Bytes {
			nbBuf = opt.HashToFieldFn.Size()
		}
		var res fr.Element
		res.SetBytes(hashBts[:nbBuf])
		publicWitness = append(publicWitness, res)
		copy(commitmentsSerialized[i*fr.Bytes:], res.Marshal())
	}

	if folded, err := pedersen.FoldCommitments(proof.Commitments, commitmentsSerialized); err != nil {
		return err
	} else {
		if err = vk.CommitmentKey.Verify(folded, proof.CommitmentPok); err != nil {
			return err
		}
	}

	// compute e(Σx.[Kvk(t)]1, -[γ]2)
	var kSum curve.G1Jac
	if _, err := kSum.MultiExp(vk.G1.K[1:], publicWitness, ecc.MultiExpConfig{}); err != nil {
		return err
	}
	kSum.AddMixed(&vk.G1.K[0])

	for i := range proof.Commitments {
		kSum.AddMixed(&proof.Commitments[i])
	}

	var kSumAff curve.G1Affine
	kSumAff.FromJacobian(&kSum)

	right, err := curve.MillerLoop([]curve.G1Affine{kSumAff}, []curve.G2Affine{vk.G2.gammaNeg})
	if err != nil {
		return err
	}

	// wait for (eKrsδ, eArBs)
	if err := <-chDone; err != nil {
		return err
	}

	fmt.Println("E", vk.e)
	fmt.Println("right", right)
	right = curve.FinalExponentiation(&right, &doubleML)
	if !vk.e.Equal(&right) {
		return errPairingCheckFailed
	}

	log.Debug().Dur("took", time.Since(start)).Msg("verifier done")
	return nil
}

// Verify verifies a proof with given VerifyingKey and publicWitness
func VerifyFolded(proof *Proof, vk *VerifyingKey, publicWitness ...fr.Vector) error {

	//nbPublicVars := len(vk.G1.K) - len(vk.PublicAndCommitmentCommitted)

	witness := PublicWitness{}
	witness.Public = publicWitness[0]
	witness.SetStartingParameters()

	/*for i := range publicWitness {
		if len(publicWitness[i].Public) != nbPublicVars-1 {
			return fmt.Errorf("invalid witness size, got %d, expected %d (public - ONE_WIRE)", len(publicWitness[i].Public), len(vk.G1.K)-1)
		}
	}*/
	log := logger.Logger().With().Str("curve", vk.CurveID().String()).Str("backend", "groth16").Logger()
	start := time.Now()
	fmt.Println("checking proof validity")
	// check that the points in the proof are in the correct subgroup
	if !proof.isValid() {
		return errCorrectSubgroupCheckFailed
	}

	var doubleML curve.GT
	chDone := make(chan error, 1)
	fmt.Println("Folding proofs")
	foldedProof, foldingParameters, err := FoldProofs(proof, proof, vk, vk, witness, witness)
	if err != nil {
		return err
	}

	fmt.Println("preparing witnesses")
	// fold public witness
	foldedWitness := FoldedWitness{}
	foldedWitness.mu = *big.NewInt(0)
	foldedWitness.H = *make([]curve.G1Affine, 1)[0].ScalarMultiplication(&make([]curve.G1Affine, 1)[0], big.NewInt(0))
	foldedWitness.E, err = curve.Pair([]curve.G1Affine{*make([]curve.G1Affine, 1)[0].ScalarMultiplication(&make([]curve.G1Affine, 1)[0], big.NewInt(0))}, make([]curve.G2Affine, 1))
	if err != nil {
		return err
	}
	fmt.Println("folding witnesses")
	foldedWitness.foldWitnesses([]PublicWitness{witness, witness}, []FoldingParameters{*foldingParameters, *foldingParameters}, *vk, []Proof{*proof, *proof})

	// compute (eKrsδ, eArBs, ealphaBeta)
	go func() {
		fmt.Println("computing eKrsδ, eArBs")
		var errML error
		krs_times_mu := make([]curve.G1Affine, 1)[0].ScalarMultiplication(&foldedProof.Krs, &foldedWitness.mu)
		doubleML, errML = curve.MillerLoop([]curve.G1Affine{*krs_times_mu, foldedProof.Ar}, []curve.G2Affine{vk.G2.deltaNeg, foldedProof.Bs})
		chDone <- errML
		close(chDone)
	}()

	/*maxNbPublicCommitted := 0
	for _, s := range vk.PublicAndCommitmentCommitted { // iterate over commitments
		maxNbPublicCommitted = utils.Max(maxNbPublicCommitted, len(s))
	}
	commitmentsSerialized := make([]byte, len(vk.PublicAndCommitmentCommitted)*fr.Bytes)
	commitmentPrehashSerialized := make([]byte, curve.SizeOfG1AffineUncompressed+maxNbPublicCommitted*fr.Bytes)
	for j := range publicWitness {
		for i := range vk.PublicAndCommitmentCommitted { // solveCommitmentWire
			copy(commitmentPrehashSerialized, proof.Commitments[i].Marshal())
			offset := curve.SizeOfG1AffineUncompressed
			for j := range vk.PublicAndCommitmentCommitted[i] {
				copy(commitmentPrehashSerialized[offset:], publicWitness[j].Public[vk.PublicAndCommitmentCommitted[i][j]-1].Marshal())
				offset += fr.Bytes
			}
			opt.HashToFieldFn.Write(commitmentPrehashSerialized[:offset])
			hashBts := opt.HashToFieldFn.Sum(nil)
			opt.HashToFieldFn.Reset()
			nbBuf := fr.Bytes
			if opt.HashToFieldFn.Size() < fr.Bytes {
				nbBuf = opt.HashToFieldFn.Size()
			}
			var res fr.Element
			res.SetBytes(hashBts[:nbBuf])
			publicWitness[j].Public = append(publicWitness[j].Public, res)
			copy(commitmentsSerialized[i*fr.Bytes:], res.Marshal())
		}
	}

	if folded, err := pedersen.FoldCommitments(proof.Commitments, commitmentsSerialized); err != nil {
		return err
	} else {
		if err = vk.CommitmentKey.Verify(folded, proof.CommitmentPok); err != nil {
			return err
		}
	}*/

	// compute e(Σx.[Kvk(t)]1, -[γ]2)
	/*
	var kSum curve.G1Jac
	if _, err := kSum.MultiExp(vk.G1.K[1:], publicWitness.Public, ecc.MultiExpConfig{}); err != nil {
		return err
	}
	kSum.AddMixed(&vk.G1.K[0])

	for i := range proof.Commitments {
		kSum.AddMixed(&proof.Commitments[i])
	}

	var kSumAff curve.G1Affine
	kSumAff.FromJacobian(&kSum)*/

	fmt.Println("computing e(Σx.[Kvk(t)]1, -[γ]2)")
	gamma_neg_times_mu := make([]curve.G2Affine, 1)[0].ScalarMultiplication(&vk.G2.gammaNeg, &foldedWitness.mu)
	right, err := curve.MillerLoop([]curve.G1Affine{foldedWitness.H}, []curve.G2Affine{*gamma_neg_times_mu})
	if err != nil {
		return err
	}

	// wait for (eKrsδ, eArBs)
	if err := <-chDone; err != nil {
		return err
	}

	right = curve.FinalExponentiation(&right, &doubleML)

	// vk.e is e(α, β), we want e(α, β)^{-mu^2}
	mu_sqrd := new(big.Int).Mul(&foldedWitness.mu, &foldedWitness.mu)
	vk.e.Exp(vk.e, mu_sqrd)
	vk.e.Inverse(&vk.e)

	vk.e.Mul(&right, &vk.e)

	if !foldedWitness.E.Equal(&vk.e) {
		return errPairingCheckFailed
	}

	log.Debug().Dur("took", time.Since(start)).Msg("verifier done")
	return nil
}

// ExportSolidity not implemented for BLS12-377
func (vk *VerifyingKey) ExportSolidity(w io.Writer) error {
	return errors.New("not implemented")
}

func FoldProofs(proof1, proof2 *Proof, vk1, vk2 *VerifyingKey, publicWitness1, publicWitness2 PublicWitness) (*FoldedProof, *FoldingParameters, error) {
	fmt.Println("eArBs")
	A1B2, err := curve.Pair([]curve.G1Affine{proof1.Ar}, []curve.G2Affine{proof2.Bs})
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("eArBs2")
	A2B1, err := curve.Pair([]curve.G1Affine{proof2.Ar}, []curve.G2Affine{proof1.Bs})
	if err != nil {
		return nil, nil, err
	}
	C1C2 := make([]curve.G1Affine, 1)[0].Add(
		make([]curve.G1Affine, 1)[0].ScalarMultiplication(&proof2.Krs, &publicWitness1.mu),
		make([]curve.G1Affine, 1)[0].ScalarMultiplication(&proof1.Krs, &publicWitness2.mu),
	)
	fmt.Println("eC1C2")
	C1C2d, err := curve.Pair([]curve.G1Affine{*C1C2}, []curve.G2Affine{vk1.G2.deltaNeg})
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("eKrsδ1")
	fmt.Println(len(vk1.G1.K[1:]))
	fmt.Println(len(publicWitness1.Public))
	var kSum1 curve.G1Jac
	if _, err := kSum1.MultiExp(vk1.G1.K[1:], publicWitness1.Public, ecc.MultiExpConfig{}); err != nil {
		return nil, nil, err
	}
	kSum1.AddMixed(&vk1.G1.K[0])
	for i := range proof1.Commitments {
		kSum1.AddMixed(&proof1.Commitments[i])
	}
	var kSumAff1 curve.G1Affine
	kSumAff1.FromJacobian(&kSum1)

	fmt.Println("eKrsδ2")
	var kSum2 curve.G1Jac
	if _, err := kSum2.MultiExp(vk2.G1.K[1:], publicWitness2.Public, ecc.MultiExpConfig{}); err != nil {
		return nil, nil, err
	}
	kSum2.AddMixed(&vk2.G1.K[0])
	for i := range proof2.Commitments {
		kSum2.AddMixed(&proof2.Commitments[i])
	}
	var kSumAff2 curve.G1Affine
	kSumAff2.FromJacobian(&kSum2)

	H1H2 := make([]curve.G1Affine, 1)[0].Add(
		make([]curve.G1Affine, 1)[0].ScalarMultiplication(&kSumAff1, &publicWitness1.mu),
		make([]curve.G1Affine, 1)[0].ScalarMultiplication(&kSumAff2, &publicWitness2.mu),
	)
	H1H2g, err := curve.Pair([]curve.G1Affine{*H1H2}, []curve.G2Affine{vk1.G2.gammaNeg})

	mu1mu2 := new(big.Int).Mul(&publicWitness1.mu, new(big.Int).Mul(&publicWitness2.mu, big.NewInt(-2)))
	emu1mu2 := make([]curve.GT, 1)[0].Exp(vk1.e, mu1mu2)

	T := A1B2.Mul(&A1B2, make([]curve.GT, 1)[0].Mul(
		&A2B1, make([]curve.GT, 1)[0].Mul(
			&C1C2d, make([]curve.GT, 1)[0].Mul(
				&H1H2g, emu1mu2))))

	r := big.NewInt(12345)				// THIS SHOULD BE RANDOM!!! Fiat shamir?
	
	//Compute the updated proof
	foldedProof := &FoldedProof{}
	foldedProof.Ar = *make([]curve.G1Affine, 1)[0].Add(&proof1.Ar, make([]curve.G1Affine, 1)[0].ScalarMultiplication(&proof2.Ar, r))
	foldedProof.Bs = *make([]curve.G2Affine, 1)[0].Add(&proof1.Bs, make([]curve.G2Affine, 1)[0].ScalarMultiplication(&proof2.Bs, r))
	foldedProof.Krs = *make([]curve.G1Affine, 1)[0].Add(&proof1.Krs, make([]curve.G1Affine, 1)[0].ScalarMultiplication(&proof2.Krs, r))

	foldingPars := &FoldingParameters{}
	foldingPars.T = *T
	foldingPars.R = *r

	return foldedProof, foldingPars, nil
}