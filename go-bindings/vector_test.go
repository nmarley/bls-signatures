package blschia_test

import (
	"bytes"
	"fmt"
	"testing"

	bls "github.com/nmarley/bls-signatures/go-bindings"
)

// Implements test vectors here:
// https://github.com/Chia-Network/bls-signatures/blob/master/SPEC.md

func TestKeygen(t *testing.T) {
	tests := []struct {
		seed          []byte
		secretKey     []byte
		pkFingerprint uint32
	}{
		{
			seed:          []byte{1, 2, 3, 4, 5},
			secretKey:     sk1Bytes,
			pkFingerprint: 0x26d53247,
		},
		{
			seed:          []byte{1, 2, 3, 4, 5, 6},
			secretKey:     []byte{},
			pkFingerprint: 0x289bb56e,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(st *testing.T) {
			sk := bls.PrivateKeyFromSeed(tt.seed)
			if len(tt.secretKey) > 0 {
				got := sk.Serialize()
				if !bytes.Equal(got, tt.secretKey) {
					st.Errorf("expected %v, got %v", tt.secretKey, got)
				}
			}

			pk := sk.PublicKey()
			fingerprint := pk.Fingerprint()
			if fingerprint != tt.pkFingerprint {
				st.Errorf("expected %v, got %v", tt.pkFingerprint, fingerprint)
			}
		})
	}
}

// Implement test for test vector for Signatures#sign
func TestVectorSignaturesSign(t *testing.T) {
	tests := []struct {
		payload     []byte
		secretKey   []byte
		expectedSig []byte
	}{
		{
			payload:     []byte{7, 8, 9},
			secretKey:   sk1Bytes,
			expectedSig: sig1Bytes,
		},
		{
			payload:     []byte{7, 8, 9},
			secretKey:   sk2Bytes,
			expectedSig: sig2Bytes,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(st *testing.T) {
			key, _ := bls.PrivateKeyFromBytes(tt.secretKey, true)
			sig := key.SignInsecure(tt.payload)
			sigBytes := sig.Serialize()
			if !bytes.Equal(sigBytes, tt.expectedSig) {
				st.Errorf("expected %v, got %v", tt.expectedSig, sigBytes)
			}
		})
	}
}

// Implement test for test vector for Signatures#verify
func TestVectorSignaturesVerify(t *testing.T) {
	tests := []struct {
		payload   []byte
		publicKey []byte
		signature []byte
	}{
		{
			payload:   payload,
			publicKey: pk1Bytes,
			signature: sig1Bytes,
		},
		{
			payload:   payload,
			publicKey: pk2Bytes,
			signature: sig2Bytes,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(st *testing.T) {
			pk, _ := bls.PublicKeyFromBytes(tt.publicKey)
			sig, _ := bls.SignatureFromBytes(tt.signature)
			sig.SetAggregationInfo(bls.AggregationInfoFromMsg(pk, tt.payload))
			if !sig.Verify() {
				st.Errorf("signature did not verify")
			}
		})
	}
}

// Implement test for test vector for HDKeys
func TestVectorHDKeys(t *testing.T) {
	seed := []byte{1, 50, 6, 244, 24, 199, 1, 25}
	pkFingerprint := uint32(0xa4700b27)
	chainCodeBytes := []byte{
		0xd8, 0xb1, 0x25, 0x55, 0xb4, 0xcc, 0x55, 0x78,
		0x95, 0x1e, 0x4a, 0x7c, 0x80, 0x03, 0x1e, 0x22,
		0x01, 0x9c, 0xc0, 0xdc, 0xe1, 0x68, 0xb3, 0xed,
		0x88, 0x11, 0x53, 0x11, 0xb8, 0xfe, 0xb1, 0xe3,
	}

	esk := bls.ExtendedPrivateKeyFromSeed(seed)
	eskFingerprint := esk.GetPublicKey().Fingerprint()
	if eskFingerprint != pkFingerprint {
		t.Errorf("got %v, expected %v", eskFingerprint, pkFingerprint)
	}

	eskCCBytes := esk.GetChainCode().Serialize()
	if !bytes.Equal(eskCCBytes, chainCodeBytes) {
		t.Errorf("got %v, expected %v", eskCCBytes, chainCodeBytes)
	}

	esk77 := esk.PrivateChild(77 + (1 << 31))
	esk77FP := esk77.GetPublicKey().Fingerprint()
	esk77FPExpected := uint32(0xa8063dcf)
	if esk77FP != esk77FPExpected {
		t.Errorf("got %v, expected %v", esk77FP, esk77FPExpected)
	}

	esk77CCBytes := esk77.GetChainCode().Serialize()
	esk77CCExpected := []byte{
		0xf2, 0xc8, 0xe4, 0x26, 0x9b, 0xb3, 0xe5, 0x4f,
		0x81, 0x79, 0xa5, 0xc6, 0x97, 0x6d, 0x92, 0xca,
		0x14, 0xc3, 0x26, 0x0d, 0xd7, 0x29, 0x98, 0x1e,
		0x9d, 0x15, 0xf5, 0x30, 0x49, 0xfd, 0x69, 0x8b,
	}
	if !bytes.Equal(esk77CCBytes, esk77CCExpected) {
		t.Errorf("got %v, expected %v", esk77CCBytes, esk77CCExpected)
	}

	fp317 := esk.PrivateChild(3).PrivateChild(17).GetPublicKey().Fingerprint()
	fp317Expected := uint32(0xff26a31f)
	if fp317 != fp317Expected {
		t.Errorf("got %v, expected %v", fp317, fp317Expected)
	}

	pubFp317 := esk.GetExtendedPublicKey().PublicChild(3).PublicChild(17).GetPublicKey().Fingerprint()
	pubFp317Expected := uint32(0xff26a31f)
	if pubFp317 != pubFp317Expected {
		t.Errorf("got %v, expected %v", pubFp317, pubFp317Expected)
	}
}

func TestVectorAggregation(t *testing.T) {
	testSigs := [][]byte{sig1Bytes, sig2Bytes}
	testPubkeys := [][]byte{pk1Bytes, pk2Bytes}
	testPayloads := [][]byte{payload, payload}

	sigs := make([]bls.Signature, len(testSigs))
	for i, sigBytes := range testSigs {
		pub, _ := bls.PublicKeyFromBytes(testPubkeys[i])
		aggInfo := bls.AggregationInfoFromMsg(pub, testPayloads[i])
		signature, _ := bls.SignatureFromBytes(sigBytes)
		signature.SetAggregationInfo(aggInfo)
		sigs[i] = signature
	}
	aggSig, _ := bls.SignatureAggregate(sigs)
	aggSigBytes := aggSig.Serialize()
	aggSigExpected := []byte{
		0x0a, 0x63, 0x84, 0x95, 0xc1, 0x40, 0x3b, 0x25,
		0xbe, 0x39, 0x1e, 0xd4, 0x4c, 0x0a, 0xb0, 0x13,
		0x39, 0x00, 0x26, 0xb5, 0x89, 0x2c, 0x79, 0x6a,
		0x85, 0xed, 0xe4, 0x63, 0x10, 0xff, 0x7d, 0x0e,
		0x06, 0x71, 0xf8, 0x6e, 0xbe, 0x0e, 0x8f, 0x56,
		0xbe, 0xe8, 0x0f, 0x28, 0xeb, 0x6d, 0x99, 0x9c,
		0x0a, 0x41, 0x8c, 0x5f, 0xc5, 0x2d, 0xeb, 0xac,
		0x8f, 0xc3, 0x38, 0x78, 0x4c, 0xd3, 0x2b, 0x76,
		0x33, 0x8d, 0x62, 0x9d, 0xc2, 0xb4, 0x04, 0x5a,
		0x58, 0x33, 0xa3, 0x57, 0x80, 0x97, 0x95, 0xef,
		0x55, 0xee, 0x3e, 0x9b, 0xee, 0x53, 0x2e, 0xdf,
		0xc1, 0xd9, 0xc4, 0x43, 0xbf, 0x5b, 0xc6, 0x58,
	}
	if !bytes.Equal(aggSigBytes, aggSigExpected) {
		t.Errorf("got %v, expected %v", aggSigBytes, aggSigExpected)
	}

	sk1, _ := bls.PrivateKeyFromBytes(sk1Bytes, true)
	sk2, _ := bls.PrivateKeyFromBytes(sk2Bytes, true)

	sig3 := sk1.Sign([]byte{1, 2, 3})
	sig4 := sk1.Sign([]byte{1, 2, 3, 4})
	sig5 := sk2.Sign([]byte{1, 2})

	aggSig2, _ := bls.SignatureAggregate([]bls.Signature{sig3, sig4, sig5})
	aggSig2Bytes := aggSig2.Serialize()
	aggSig2Expected := []byte{
		0x8b, 0x11, 0xda, 0xf7, 0x3c, 0xd0, 0x5f, 0x2f,
		0xe2, 0x78, 0x09, 0xb7, 0x4a, 0x7b, 0x4c, 0x65,
		0xb1, 0xbb, 0x79, 0xcc, 0x10, 0x66, 0xbd, 0xf8,
		0x39, 0xd9, 0x6b, 0x97, 0xe0, 0x73, 0xc1, 0xa6,
		0x35, 0xd2, 0xec, 0x04, 0x8e, 0x08, 0x01, 0xb4,
		0xa2, 0x08, 0x11, 0x8f, 0xdb, 0xbb, 0x63, 0xa5,
		0x16, 0xba, 0xb8, 0x75, 0x5c, 0xc8, 0xd8, 0x50,
		0x86, 0x2e, 0xea, 0xa0, 0x99, 0x54, 0x0c, 0xd8,
		0x36, 0x21, 0xff, 0x9d, 0xb9, 0x7b, 0x4a, 0xda,
		0x85, 0x7e, 0xf5, 0x4c, 0x50, 0x71, 0x54, 0x86,
		0x21, 0x7b, 0xd2, 0xec, 0xb4, 0x51, 0x7e, 0x05,
		0xab, 0x49, 0x38, 0x0c, 0x04, 0x1e, 0x15, 0x9b,
	}
	if !bytes.Equal(aggSig2Bytes, aggSig2Expected) {
		t.Errorf("got %v, expected %v", aggSig2Bytes, aggSig2Expected)
	}
	if !aggSig2.Verify() {
		t.Errorf("aggSig2 did not verify")
	}

	sig1, _ := bls.SignatureFromBytes(sig1Bytes)
	sig2, _ := bls.SignatureFromBytes(sig2Bytes)
	pk1, _ := bls.PublicKeyFromBytes(pk1Bytes)
	pk2, _ := bls.PublicKeyFromBytes(pk2Bytes)
	mh := Sha256(payload)

	sig1.SetAggregationInfo(bls.AggregationInfoFromMsgHash(pk1, mh))
	sig2.SetAggregationInfo(bls.AggregationInfoFromMsgHash(pk2, mh))

	aggPk, err := bls.PublicKeyAggregate([]bls.PublicKey{pk1, pk2})
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	aggPkBytes := aggPk.Serialize()
	aggPkExpected := []byte{
		0x13, 0xff, 0x74, 0xea, 0x55, 0x95, 0x29, 0x24,
		0xe8, 0x24, 0xc5, 0xa0, 0x88, 0x25, 0xe3, 0xc3,
		0x6d, 0x92, 0x8d, 0xf7, 0xfb, 0xa1, 0x5b, 0xf4,
		0x92, 0xd0, 0x0a, 0x6a, 0x11, 0x28, 0x68, 0x62,
		0x5f, 0x77, 0x2c, 0x91, 0x02, 0xf2, 0xd9, 0xe2,
		0x1b, 0x99, 0xbf, 0x99, 0xfd, 0xc6, 0x27, 0xb6,
	}
	if !bytes.Equal(aggPkBytes, aggPkExpected) {
		t.Errorf("got %v, expected %v", aggPkBytes, aggPkExpected)
	}

	toMergeAIs := []bls.AggregationInfo{
		sig1.GetAggregationInfo(),
		sig2.GetAggregationInfo(),
	}
	ai := bls.MergeAggregationInfos(toMergeAIs)
	aggSig2.SetAggregationInfo(ai)
	if !aggSig.Verify() {
		t.Errorf("aggSig did not verify")
	}
	mh = []byte{7, 8, 9}
	ai = bls.AggregationInfoFromMsgHash(pk2, mh)
	sig1.SetAggregationInfo(ai)
	if sig1.Verify() {
		t.Errorf("sig1 should have not have verified")
	}

	toMergeAIs = []bls.AggregationInfo{
		sig3.GetAggregationInfo(),
		sig4.GetAggregationInfo(),
		sig5.GetAggregationInfo(),
	}
	aggSig2.SetAggregationInfo(bls.MergeAggregationInfos(toMergeAIs))
	if !aggSig2.Verify() {
		t.Errorf("aggSig2 did not verify")
	}
}

func TestVectorAggregation2(t *testing.T) {
	m1 := []byte{1, 2, 3, 40}
	m2 := []byte{5, 6, 70, 201}
	m3 := []byte{9, 10, 11, 12, 13}
	m4 := []byte{15, 63, 244, 92, 0, 1}

	sk1 := bls.PrivateKeyFromSeed([]byte{1, 2, 3, 4, 5})
	sk2 := bls.PrivateKeyFromSeed([]byte{1, 2, 3, 4, 5, 6})

	sig1 := sk1.Sign(m1)
	sig2 := sk2.Sign(m2)
	sig3 := sk2.Sign(m1)
	sig4 := sk1.Sign(m3)
	sig5 := sk1.Sign(m1)
	sig6 := sk1.Sign(m4)

	sigL, _ := bls.SignatureAggregate([]bls.Signature{sig1, sig2})
	sigR, _ := bls.SignatureAggregate([]bls.Signature{sig3, sig4, sig5})
	if !sigL.Verify() {
		t.Errorf("sigL did not verify")
	}
	if !sigR.Verify() {
		t.Errorf("sigR did not verify")
	}

	sigFinal, _ := bls.SignatureAggregate([]bls.Signature{sigL, sigR, sig6})
	sigFinalBytes := sigFinal.Serialize()
	sigFinalExpected := []byte{
		0x07, 0x96, 0x99, 0x58, 0xfb, 0xf8, 0x2e, 0x65,
		0xbd, 0x13, 0xba, 0x07, 0x49, 0x99, 0x07, 0x64,
		0xca, 0xc8, 0x1c, 0xf1, 0x0d, 0x92, 0x3a, 0xf9,
		0xfd, 0xd2, 0x72, 0x3f, 0x1e, 0x39, 0x10, 0xc3,
		0xfd, 0xb8, 0x74, 0xa6, 0x7f, 0x9d, 0x51, 0x1b,
		0xb7, 0xe4, 0x92, 0x0f, 0x8c, 0x01, 0x23, 0x2b,
		0x12, 0xe2, 0xfb, 0x5e, 0x64, 0xa7, 0xc2, 0xd1,
		0x77, 0xa4, 0x75, 0xda, 0xb5, 0xc3, 0x72, 0x9c,
		0xa1, 0xf5, 0x80, 0x30, 0x1c, 0xcd, 0xef, 0x80,
		0x9c, 0x57, 0xa8, 0x84, 0x68, 0x90, 0x26, 0x5d,
		0x19, 0x5b, 0x69, 0x4f, 0xa4, 0x14, 0xa2, 0xa3,
		0xaa, 0x55, 0xc3, 0x28, 0x37, 0xfd, 0xdd, 0x80,
	}
	if !bytes.Equal(sigFinalBytes, sigFinalExpected) {
		t.Errorf("got %v, expected %v", sigFinalBytes, sigFinalExpected)
	}
	if !sigFinal.Verify() {
		t.Errorf("sigFinal did not verify")
	}

	// Begin Signature Division

	quotient, _ := sigFinal.DivideBy([]bls.Signature{sig2, sig5, sig6})
	quotientBytes := quotient.Serialize()
	quotientExpected := []byte{
		0x8e, 0xbc, 0x8a, 0x73, 0xa2, 0x29, 0x1e, 0x68,
		0x9c, 0xe5, 0x17, 0x69, 0xff, 0x87, 0xe5, 0x17,
		0xbe, 0x60, 0x89, 0xfd, 0x06, 0x27, 0xb2, 0xce,
		0x3c, 0xd2, 0xf0, 0xee, 0x1c, 0xe1, 0x34, 0xb3,
		0x9c, 0x4d, 0xa4, 0x09, 0x28, 0x95, 0x41, 0x75,
		0x01, 0x4e, 0x9b, 0xbe, 0x62, 0x3d, 0x84, 0x5d,
		0x0b, 0xdb, 0xa8, 0xbf, 0xd2, 0xa8, 0x5a, 0xf9,
		0x50, 0x7d, 0xdf, 0x14, 0x55, 0x79, 0x48, 0x01,
		0x32, 0xb6, 0x76, 0xf0, 0x27, 0x38, 0x13, 0x14,
		0xd9, 0x83, 0xa6, 0x38, 0x42, 0xfc, 0xc7, 0xbf,
		0x5c, 0x8c, 0x08, 0x84, 0x61, 0xe3, 0xeb, 0xb0,
		0x4d, 0xcf, 0x86, 0xb4, 0x31, 0xd6, 0x23, 0x8f,
	}
	if !bytes.Equal(quotientBytes, quotientExpected) {
		t.Errorf("got %v, expected %v", quotientBytes, quotientExpected)
	}
	if !quotient.Verify() {
		t.Errorf("quotient did not verify")
	}

	// Ensure that dividing by an empty list returns the same signature
	quotEmptyDiv, _ := quotient.DivideBy([]bls.Signature{})
	if !quotEmptyDiv.Equal(quotient) {
		t.Errorf("got %v, expected %v", quotEmptyDiv, quotient)
	}

	// should throw with not a subset
	_, err := quotient.DivideBy([]bls.Signature{sig6})
	if err == nil {
		t.Error("did not get expected error")
	}

	// should NOT throw
	_, err = sigFinal.DivideBy([]bls.Signature{sig1})
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}

	// should throw with not unique error
	_, err = sigFinal.DivideBy([]bls.Signature{sigL})
	if err == nil {
		t.Error("did not get expected error")
	}

	// Divide by aggregate
	sig7 := sk2.Sign(m3)
	sig8 := sk2.Sign(m4)

	sigR2, _ := bls.SignatureAggregate([]bls.Signature{sig7, sig8})
	sigFinal2, _ := bls.SignatureAggregate([]bls.Signature{sigFinal, sigR2})

	quotient2, _ := sigFinal2.DivideBy([]bls.Signature{sigR2})
	quotient2Bytes := quotient2.Serialize()
	quotient2Expected := []byte{
		0x06, 0xaf, 0x69, 0x30, 0xbd, 0x06, 0x83, 0x8f,
		0x2e, 0x4b, 0x00, 0xb6, 0x29, 0x11, 0xfb, 0x29,
		0x02, 0x45, 0xcc, 0xe5, 0x03, 0xcc, 0xf5, 0xbf,
		0xc2, 0x90, 0x14, 0x59, 0x89, 0x77, 0x31, 0xdd,
		0x08, 0xfc, 0x4c, 0x56, 0xdb, 0xde, 0x75, 0xa1,
		0x16, 0x77, 0xcc, 0xfb, 0xfa, 0x61, 0xab, 0x8b,
		0x14, 0x73, 0x5f, 0xdd, 0xc6, 0x6a, 0x02, 0xb7,
		0xae, 0xeb, 0xb5, 0x4a, 0xb9, 0xa4, 0x14, 0x88,
		0xf8, 0x9f, 0x64, 0x1d, 0x83, 0xd4, 0x51, 0x5c,
		0x4d, 0xd2, 0x0d, 0xfc, 0xf2, 0x8c, 0xbb, 0xcc,
		0xb1, 0x47, 0x2c, 0x32, 0x7f, 0x07, 0x80, 0xbe,
		0x3a, 0x90, 0xc0, 0x05, 0xc5, 0x8a, 0x47, 0xd3,
	}
	if !bytes.Equal(quotient2Bytes, quotient2Expected) {
		t.Errorf("got %v, expected %v", quotient2Bytes, quotient2Expected)
	}
	if !quotient2.Verify() {
		t.Errorf("quotient2 did not verify")
	}
}

// Values either defined in or derived from test vectors and re-used multiple
// times
var sk1Bytes = []byte{
	0x02, 0x2f, 0xb4, 0x2c, 0x08, 0xc1, 0x2d, 0xe3,
	0xa6, 0xaf, 0x05, 0x38, 0x80, 0x19, 0x98, 0x06,
	0x53, 0x2e, 0x79, 0x51, 0x5f, 0x94, 0xe8, 0x34,
	0x61, 0x61, 0x21, 0x01, 0xf9, 0x41, 0x2f, 0x9e,
}
var sk2Bytes = []byte{
	0x50, 0x2c, 0x56, 0x61, 0xf5, 0xaf, 0x46, 0xed,
	0x48, 0xdd, 0xc3, 0xb5, 0x33, 0x2e, 0x21, 0xb9,
	0x3c, 0xc7, 0xd0, 0xa8, 0x4d, 0xf4, 0x6c, 0x4b,
	0x9c, 0x7f, 0xe8, 0xf2, 0x5e, 0xf4, 0x8d, 0x66,
}

var sig1Bytes = []byte{
	0x93, 0xeb, 0x2e, 0x1c, 0xb5, 0xef, 0xcf, 0xb3,
	0x1f, 0x2c, 0x08, 0xb2, 0x35, 0xe8, 0x20, 0x3a,
	0x67, 0x26, 0x5b, 0xc6, 0xa1, 0x3d, 0x9f, 0x0a,
	0xb7, 0x77, 0x27, 0x29, 0x3b, 0x74, 0xa3, 0x57,
	0xff, 0x04, 0x59, 0xac, 0x21, 0x0d, 0xc8, 0x51,
	0xfc, 0xb8, 0xa6, 0x0c, 0xb7, 0xd3, 0x93, 0xa4,
	0x19, 0x91, 0x5c, 0xfc, 0xf8, 0x39, 0x08, 0xdd,
	0xbe, 0xac, 0x32, 0x03, 0x9a, 0xaa, 0x3e, 0x8f,
	0xea, 0x82, 0xef, 0xcb, 0x3b, 0xa4, 0xf7, 0x40,
	0xf2, 0x0c, 0x76, 0xdf, 0x5e, 0x97, 0x10, 0x9b,
	0x57, 0x37, 0x0a, 0xe3, 0x2d, 0x9b, 0x70, 0xd2,
	0x56, 0xa9, 0x89, 0x42, 0xe5, 0x80, 0x60, 0x65,
}
var sig2Bytes = []byte{
	0x97, 0x5b, 0x5d, 0xaa, 0x64, 0xb9, 0x15, 0xbe,
	0x19, 0xb5, 0xac, 0x6d, 0x47, 0xbc, 0x1c, 0x2f,
	0xc8, 0x32, 0xd2, 0xfb, 0x8c, 0xa3, 0xe9, 0x5c,
	0x48, 0x05, 0xd8, 0x21, 0x6f, 0x95, 0xcf, 0x2b,
	0xdb, 0xb3, 0x6c, 0xc2, 0x36, 0x45, 0xf5, 0x20,
	0x40, 0xe3, 0x81, 0x55, 0x07, 0x27, 0xdb, 0x42,
	0x0b, 0x52, 0x3b, 0x57, 0xd4, 0x94, 0x95, 0x9e,
	0x0e, 0x8c, 0x0c, 0x60, 0x60, 0xc4, 0x6c, 0xf1,
	0x73, 0x87, 0x28, 0x97, 0xf1, 0x4d, 0x43, 0xb2,
	0xac, 0x2a, 0xec, 0x52, 0xfc, 0x7b, 0x46, 0xc0,
	0x2c, 0x56, 0x99, 0xff, 0x7a, 0x10, 0xbe, 0xba,
	0x24, 0xd3, 0xce, 0xd4, 0xe8, 0x9c, 0x82, 0x1e,
}
var payload = []byte{7, 8, 9}

var pk1Bytes = []byte{
	0x02, 0xa8, 0xd2, 0xaa, 0xa6, 0xa5, 0xe2, 0xe0,
	0x8d, 0x4b, 0x8d, 0x40, 0x6a, 0xaf, 0x01, 0x21,
	0xa2, 0xfc, 0x20, 0x88, 0xed, 0x12, 0x43, 0x1e,
	0x6b, 0x06, 0x63, 0x02, 0x8d, 0xa9, 0xac, 0x59,
	0x22, 0xc9, 0xea, 0x91, 0xcd, 0xe7, 0xdd, 0x74,
	0xb7, 0xd7, 0x95, 0x58, 0x0a, 0xcc, 0x7a, 0x61,
}
var pk2Bytes = []byte{
	0x83, 0xfb, 0xcb, 0xbf, 0xa6, 0xb7, 0xa5, 0xa0,
	0xe7, 0x07, 0xef, 0xaa, 0x9e, 0x6d, 0xe2, 0x58,
	0xa7, 0x9a, 0x59, 0x11, 0x6d, 0xd8, 0x89, 0xce,
	0x74, 0xf1, 0xab, 0x7f, 0x54, 0xc9, 0xb7, 0xba,
	0x15, 0x43, 0x9d, 0xcb, 0x4a, 0xcf, 0xbd, 0xd8,
	0xbc, 0xff, 0xdd, 0x88, 0x25, 0x79, 0x5b, 0x90,
}