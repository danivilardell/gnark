// Copyright 2020 ConsenSys AG
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

// Package groth16 implements Groth16 Zero Knowledge Proof system  (aka zkSNARK).
//
// # See also
//
// https://eprint.iacr.org/2016/260.pdf
package groth16

import (
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	cs_bls12377 "github.com/consensys/gnark/constraint/bls12-377"
	cs_bls12381 "github.com/consensys/gnark/constraint/bls12-381"
	cs_bls24315 "github.com/consensys/gnark/constraint/bls24-315"
	cs_bls24317 "github.com/consensys/gnark/constraint/bls24-317"
	cs_bn254 "github.com/consensys/gnark/constraint/bn254"
	cs_bw6633 "github.com/consensys/gnark/constraint/bw6-633"
	cs_bw6761 "github.com/consensys/gnark/constraint/bw6-761"

	fr_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	fr_bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	fr_bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	fr_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	fr_bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	fr_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"

	gnarkio "github.com/consensys/gnark/io"

	groth16_bls12377 "github.com/consensys/gnark/backend/groth16/bls12-377"
	groth16_bls12381 "github.com/consensys/gnark/backend/groth16/bls12-381"
	groth16_bls24315 "github.com/consensys/gnark/backend/groth16/bls24-315"
	groth16_bls24317 "github.com/consensys/gnark/backend/groth16/bls24-317"
	groth16_bn254 "github.com/consensys/gnark/backend/groth16/bn254"
	icicle_bn254 "github.com/consensys/gnark/backend/groth16/bn254/icicle"
	groth16_bw6633 "github.com/consensys/gnark/backend/groth16/bw6-633"
	groth16_bw6761 "github.com/consensys/gnark/backend/groth16/bw6-761"

	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"
)

type groth16Object interface {
	gnarkio.WriterRawTo
	io.WriterTo
	io.ReaderFrom
	CurveID() ecc.ID
}

// Proof represents a Groth16 proof generated by groth16.Prove
//
// it's underlying implementation is curve specific (see gnark/internal/backend)
type Proof interface {
	groth16Object
}

type FoldedProof interface {
}

type FoldingParameters interface {
}

// ProvingKey represents a Groth16 ProvingKey
//
// it's underlying implementation is strongly typed with the curve (see gnark/internal/backend)
type ProvingKey interface {
	groth16Object
	gnarkio.UnsafeReaderFrom

	// NbG1 returns the number of G1 elements in the ProvingKey
	NbG1() int

	// NbG2 returns the number of G2 elements in the ProvingKey
	NbG2() int

	IsDifferent(interface{}) bool
}

// VerifyingKey represents a Groth16 VerifyingKey
//
// it's underlying implementation is strongly typed with the curve (see gnark/internal/backend)
//
// ExportSolidity is implemented for BN254 and will return an error with other curves
type VerifyingKey interface {
	groth16Object
	gnarkio.UnsafeReaderFrom

	// NbPublicWitness returns number of elements expected in the public witness
	NbPublicWitness() int

	// NbG1 returns the number of G1 elements in the VerifyingKey
	NbG1() int

	// NbG2 returns the number of G2 elements in the VerifyingKey
	NbG2() int

	// ExportSolidity writes a solidity Verifier contract from the VerifyingKey
	// this will return an error if not supported on the CurveID()
	ExportSolidity(w io.Writer) error

	IsDifferent(interface{}) bool
}

// Verify runs the groth16.Verify algorithm on provided proof with given witness
func Verify(proof Proof, vk VerifyingKey, publicWitness witness.Witness, opts ...backend.VerifierOption) error {

	switch _proof := proof.(type) {
	case *groth16_bls12377.Proof:
		w, ok := publicWitness.Vector().(fr_bls12377.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bls12377.Verify(_proof, vk.(*groth16_bls12377.VerifyingKey), w, opts...)
	case *groth16_bls12381.Proof:
		w, ok := publicWitness.Vector().(fr_bls12381.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bls12381.Verify(_proof, vk.(*groth16_bls12381.VerifyingKey), w, opts...)
	case *groth16_bn254.Proof:
		w, ok := publicWitness.Vector().(fr_bn254.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bn254.Verify(_proof, vk.(*groth16_bn254.VerifyingKey), w, opts...)
	case *groth16_bw6761.Proof:
		w, ok := publicWitness.Vector().(fr_bw6761.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bw6761.Verify(_proof, vk.(*groth16_bw6761.VerifyingKey), w, opts...)
	case *groth16_bls24317.Proof:
		w, ok := publicWitness.Vector().(fr_bls24317.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bls24317.Verify(_proof, vk.(*groth16_bls24317.VerifyingKey), w, opts...)
	case *groth16_bls24315.Proof:
		w, ok := publicWitness.Vector().(fr_bls24315.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bls24315.Verify(_proof, vk.(*groth16_bls24315.VerifyingKey), w, opts...)
	case *groth16_bw6633.Proof:
		w, ok := publicWitness.Vector().(fr_bw6633.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return groth16_bw6633.Verify(_proof, vk.(*groth16_bw6633.VerifyingKey), w, opts...)
	default:
		panic("unrecognized R1CS curve type")
	}
}

// FIX FOLDING PARAMETERS FOR MORE THAN 2 PROOFS FOLDED
func VerifyFolded(proof FoldedProof, foldingParameters []FoldingParameters, vk VerifyingKey, publicWitness []witness.Witness, proofs []Proof) error {
	proofs_bls12377 := make([]groth16_bls12377.Proof, len(proofs))
	for i, _proof := range proofs {
		proofs_bls12377[i] = *_proof.(*groth16_bls12377.Proof)
	}
	witness_bls12377 := make([]fr_bls12377.Vector, len(publicWitness))
	for i, w := range publicWitness {
		witness_bls12377[i] = w.Vector().(fr_bls12377.Vector)
	}
	folding_pars_bls12377 := make([]groth16_bls12377.FoldingParameters, len(foldingParameters))
	for i, f := range foldingParameters {
		folding_pars_bls12377[i] = *f.(*groth16_bls12377.FoldingParameters)
	}
	return groth16_bls12377.VerifyFolded(proof.(*groth16_bls12377.FoldedProof), folding_pars_bls12377, vk.(*groth16_bls12377.VerifyingKey), proofs_bls12377, witness_bls12377)
}

func FoldProofs(proofs []Proof, vk VerifyingKey, opts ...backend.ProverOption) (FoldedProof, error) {
	foldedProof := &groth16_bls12377.FoldedProof{}
	proof0 := proofs[0].(*groth16_bls12377.Proof)
	foldedProof.Ar = proof0.Ar
	foldedProof.Bs = proof0.Bs
	foldedProof.Krs = proof0.Krs
	for i, _ := range proofs {
		switch _proof := proofs[i].(type) {
		case *groth16_bls12377.Proof:
			if i == 0 {
				continue
			}
			var err error
			foldedProof, err = groth16_bls12377.FoldProof(foldedProof, _proof, vk.(*groth16_bls12377.VerifyingKey))
			if err != nil {
				return nil, err
			}
		default:
			panic("unrecognized R1CS curve type")
		}
	}
	return foldedProof, nil
}

func GetFoldingParameters(proofs []Proof, vk VerifyingKey, publicWitness []witness.Witness, opts ...backend.ProverOption) ([]FoldingParameters, error) {
	w1, _ := publicWitness[0].Vector().(fr_bls12377.Vector)
	kSumAff, err := groth16_bls12377.GetkSumAff(proofs[0].(*groth16_bls12377.Proof), vk.(*groth16_bls12377.VerifyingKey), w1)
	if err != nil {
		return nil, err
	}

	foldedProof := &groth16_bls12377.FoldedProof{}
	proof0 := proofs[0].(*groth16_bls12377.Proof)
	foldedProof.Ar = proof0.Ar
	foldedProof.Bs = proof0.Bs
	foldedProof.Krs = proof0.Krs

	foldingParameters := make([]FoldingParameters, len(proofs)-1)

	foldedWitness := groth16_bls12377.FoldedWitness{}
	foldedWitness.H = kSumAff
	foldedWitness.E = *make([]curve.GT, 1)[0].SetOne()
	foldedWitness.Mu = *big.NewInt(1)

	for i, _ := range proofs {
		switch _proof := proofs[i].(type) {
		case *groth16_bls12377.Proof:
			if i == 0 {
				continue
			}
			w, _ := publicWitness[i].Vector().(fr_bls12377.Vector)
			var err error
			foldingParameters[i-1], kSumAff, err = groth16_bls12377.GetFoldingParameters(foldedWitness.H, foldedProof, proofs[i].(*groth16_bls12377.Proof), vk.(*groth16_bls12377.VerifyingKey), foldedWitness, w)
			if err != nil {
				return nil, err
			}
			witness := groth16_bls12377.PublicWitness{}
			witness.Public = w
			witness.SetStartingParameters()
			foldingPars := *foldingParameters[i-1].(*groth16_bls12377.FoldingParameters)
			foldedWitness.FoldWitnesses([]groth16_bls12377.PublicWitness{witness}, []groth16_bls12377.FoldingParameters{foldingPars}, *vk.(*groth16_bls12377.VerifyingKey), []groth16_bls12377.Proof{*proofs[i].(*groth16_bls12377.Proof)})
			foldedProof, _ = groth16_bls12377.FoldProof(foldedProof, _proof, vk.(*groth16_bls12377.VerifyingKey))
		default:
			panic("unrecognized R1CS curve type")
		}
	}
	return foldingParameters, nil
}

// Prove runs the groth16.Prove algorithm.
//
// if the force flag is set:
//
//		will execute all the prover computations, even if the witness is invalid
//	 will produce an invalid proof
//		internally, the solution vector to the R1CS will be filled with random values which may impact benchmarking
func Prove(r1cs constraint.ConstraintSystem, pk ProvingKey, fullWitness witness.Witness, opts ...backend.ProverOption) (Proof, error) {
	switch _r1cs := r1cs.(type) {
	case *cs_bls12377.R1CS:
		return groth16_bls12377.Prove(_r1cs, pk.(*groth16_bls12377.ProvingKey), fullWitness, opts...)

	case *cs_bls12381.R1CS:
		return groth16_bls12381.Prove(_r1cs, pk.(*groth16_bls12381.ProvingKey), fullWitness, opts...)

	case *cs_bn254.R1CS:
		if icicle_bn254.HasIcicle {
			return icicle_bn254.Prove(_r1cs, pk.(*icicle_bn254.ProvingKey), fullWitness, opts...)
		}
		return groth16_bn254.Prove(_r1cs, pk.(*groth16_bn254.ProvingKey), fullWitness, opts...)

	case *cs_bw6761.R1CS:
		return groth16_bw6761.Prove(_r1cs, pk.(*groth16_bw6761.ProvingKey), fullWitness, opts...)

	case *cs_bls24317.R1CS:
		return groth16_bls24317.Prove(_r1cs, pk.(*groth16_bls24317.ProvingKey), fullWitness, opts...)

	case *cs_bls24315.R1CS:
		return groth16_bls24315.Prove(_r1cs, pk.(*groth16_bls24315.ProvingKey), fullWitness, opts...)

	case *cs_bw6633.R1CS:
		return groth16_bw6633.Prove(_r1cs, pk.(*groth16_bw6633.ProvingKey), fullWitness, opts...)

	default:
		panic("unrecognized R1CS curve type")
	}
}

// Setup runs groth16.Setup with provided R1CS and outputs a key pair associated with the circuit.
//
// Note that careful consideration must be given to this step in a production environment.
// groth16.Setup uses some randomness to precompute the Proving and Verifying keys. If the process
// or machine leaks this randomness, an attacker could break the ZKP protocol.
//
// Two main solutions to this deployment issues are: running the Setup through a MPC (multi party computation)
// or using a ZKP backend like PLONK where the per-circuit Setup is deterministic.
func Setup(r1cs constraint.ConstraintSystem) (ProvingKey, VerifyingKey, error) {

	switch _r1cs := r1cs.(type) {
	case *cs_bls12377.R1CS:
		var pk groth16_bls12377.ProvingKey
		var vk groth16_bls12377.VerifyingKey
		if err := groth16_bls12377.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bls12381.R1CS:
		var pk groth16_bls12381.ProvingKey
		var vk groth16_bls12381.VerifyingKey
		if err := groth16_bls12381.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bn254.R1CS:
		var vk groth16_bn254.VerifyingKey
		if icicle_bn254.HasIcicle {
			var pk icicle_bn254.ProvingKey
			if err := icicle_bn254.Setup(_r1cs, &pk, &vk); err != nil {
				return nil, nil, err
			}
			return &pk, &vk, nil
		}
		var pk groth16_bn254.ProvingKey
		if err := groth16_bn254.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bw6761.R1CS:
		var pk groth16_bw6761.ProvingKey
		var vk groth16_bw6761.VerifyingKey
		if err := groth16_bw6761.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bls24317.R1CS:
		var pk groth16_bls24317.ProvingKey
		var vk groth16_bls24317.VerifyingKey
		if err := groth16_bls24317.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bls24315.R1CS:
		var pk groth16_bls24315.ProvingKey
		var vk groth16_bls24315.VerifyingKey
		if err := groth16_bls24315.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	case *cs_bw6633.R1CS:
		var pk groth16_bw6633.ProvingKey
		var vk groth16_bw6633.VerifyingKey
		if err := groth16_bw6633.Setup(_r1cs, &pk, &vk); err != nil {
			return nil, nil, err
		}
		return &pk, &vk, nil
	default:
		panic("unrecognized R1CS curve type")
	}
}

// DummySetup create a random ProvingKey with provided R1CS
// it doesn't return a VerifyingKey and is use for benchmarking or test purposes only.
func DummySetup(r1cs constraint.ConstraintSystem) (ProvingKey, error) {
	switch _r1cs := r1cs.(type) {
	case *cs_bls12377.R1CS:
		var pk groth16_bls12377.ProvingKey
		if err := groth16_bls12377.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bls12381.R1CS:
		var pk groth16_bls12381.ProvingKey
		if err := groth16_bls12381.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bn254.R1CS:
		if icicle_bn254.HasIcicle {
			var pk icicle_bn254.ProvingKey
			if err := icicle_bn254.DummySetup(_r1cs, &pk); err != nil {
				return nil, err
			}
			return &pk, nil
		}
		var pk groth16_bn254.ProvingKey
		if err := groth16_bn254.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bw6761.R1CS:
		var pk groth16_bw6761.ProvingKey
		if err := groth16_bw6761.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bls24317.R1CS:
		var pk groth16_bls24317.ProvingKey
		if err := groth16_bls24317.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bls24315.R1CS:
		var pk groth16_bls24315.ProvingKey
		if err := groth16_bls24315.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	case *cs_bw6633.R1CS:
		var pk groth16_bw6633.ProvingKey
		if err := groth16_bw6633.DummySetup(_r1cs, &pk); err != nil {
			return nil, err
		}
		return &pk, nil
	default:
		panic("unrecognized R1CS curve type")
	}
}

// NewProvingKey instantiates a curve-typed ProvingKey and returns an interface object
// This function exists for serialization purposes
func NewProvingKey(curveID ecc.ID) ProvingKey {
	var pk ProvingKey
	switch curveID {
	case ecc.BN254:
		pk = &groth16_bn254.ProvingKey{}
		if icicle_bn254.HasIcicle {
			pk = &icicle_bn254.ProvingKey{}
		}
	case ecc.BLS12_377:
		pk = &groth16_bls12377.ProvingKey{}
	case ecc.BLS12_381:
		pk = &groth16_bls12381.ProvingKey{}
	case ecc.BW6_761:
		pk = &groth16_bw6761.ProvingKey{}
	case ecc.BLS24_317:
		pk = &groth16_bls24317.ProvingKey{}
	case ecc.BLS24_315:
		pk = &groth16_bls24315.ProvingKey{}
	case ecc.BW6_633:
		pk = &groth16_bw6633.ProvingKey{}
	default:
		panic("not implemented")
	}
	return pk
}

// NewVerifyingKey instantiates a curve-typed VerifyingKey and returns an interface
// This function exists for serialization purposes
func NewVerifyingKey(curveID ecc.ID) VerifyingKey {
	var vk VerifyingKey
	switch curveID {
	case ecc.BN254:
		vk = &groth16_bn254.VerifyingKey{}
	case ecc.BLS12_377:
		vk = &groth16_bls12377.VerifyingKey{}
	case ecc.BLS12_381:
		vk = &groth16_bls12381.VerifyingKey{}
	case ecc.BW6_761:
		vk = &groth16_bw6761.VerifyingKey{}
	case ecc.BLS24_317:
		vk = &groth16_bls24317.VerifyingKey{}
	case ecc.BLS24_315:
		vk = &groth16_bls24315.VerifyingKey{}
	case ecc.BW6_633:
		vk = &groth16_bw6633.VerifyingKey{}
	default:
		panic("not implemented")
	}

	return vk
}

// NewProof instantiates a curve-typed Proof and returns an interface
// This function exists for serialization purposes
func NewProof(curveID ecc.ID) Proof {
	var proof Proof
	switch curveID {
	case ecc.BN254:
		proof = &groth16_bn254.Proof{}
	case ecc.BLS12_377:
		proof = &groth16_bls12377.Proof{}
	case ecc.BLS12_381:
		proof = &groth16_bls12381.Proof{}
	case ecc.BW6_761:
		proof = &groth16_bw6761.Proof{}
	case ecc.BLS24_317:
		proof = &groth16_bls24317.Proof{}
	case ecc.BLS24_315:
		proof = &groth16_bls24315.Proof{}
	case ecc.BW6_633:
		proof = &groth16_bw6633.Proof{}
	default:
		panic("not implemented")
	}

	return proof
}

// NewCS instantiate a concrete curved-typed R1CS and return a R1CS interface
// This method exists for (de)serialization purposes
func NewCS(curveID ecc.ID) constraint.ConstraintSystem {
	var r1cs constraint.ConstraintSystem
	switch curveID {
	case ecc.BN254:
		r1cs = &cs_bn254.R1CS{}
	case ecc.BLS12_377:
		r1cs = &cs_bls12377.R1CS{}
	case ecc.BLS12_381:
		r1cs = &cs_bls12381.R1CS{}
	case ecc.BW6_761:
		r1cs = &cs_bw6761.R1CS{}
	case ecc.BLS24_317:
		r1cs = &cs_bls24317.R1CS{}
	case ecc.BLS24_315:
		r1cs = &cs_bls24315.R1CS{}
	case ecc.BW6_633:
		r1cs = &cs_bw6633.R1CS{}
	default:
		panic("not implemented")
	}
	return r1cs
}
