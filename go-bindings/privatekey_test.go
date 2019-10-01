package blschia_test

import (
	"bytes"
	"math/big"
	"testing"

	bls "github.com/nmarley/bls-signatures/go-bindings"
)

func TestPrivateKey(t *testing.T) {
	// Example seed, used to generate private key. Always use
	// a secure RNG with sufficient entropy to generate a seed.
	seed := []byte{
		0x00, 0x32, 0x06, 0xf4, 0x18, 0xc7, 0x01, 0x19,
		0x34, 0x58, 0xc0, 0x13, 0x12, 0x0c, 0x59, 0x06,
		0xdc, 0x12, 0x66, 0x3a, 0xd1, 0x52, 0x0c, 0x3e,
		0x59, 0x6e, 0xb6, 0x09, 0x2c, 0x14, 0xfe, 0x16,
	}
	result := []byte{
		0x54, 0x3d, 0x7c, 0x46, 0xcb, 0xbf, 0x5b, 0xaa,
		0xbc, 0x4a, 0xb0, 0x16, 0x61, 0xfa, 0xa9, 0x1a,
		0x69, 0x3b, 0x27, 0x04, 0xf6, 0x67, 0xe3, 0x35,
		0x85, 0xe4, 0xd6, 0x3b, 0xa5, 0x06, 0x9e, 0x27,
	}
	sk := bls.PrivateKeyFromSeed(seed)
	keyBytes := sk.Serialize()
	if !bytes.Equal(keyBytes, result) {
		t.Errorf("got %v, expected %v", keyBytes, result)
	}

	pkExpected := []byte{
		0x02, 0xa7, 0xc0, 0x64, 0xfa, 0xf6, 0x1b, 0x55,
		0xd3, 0x39, 0x42, 0xa3, 0x2d, 0xe9, 0x9a, 0x00,
		0xfa, 0x96, 0x26, 0x2e, 0x5f, 0x09, 0xc7, 0x3c,
		0xbe, 0x60, 0x71, 0xe5, 0x8e, 0xbf, 0xad, 0x66,
		0xdb, 0x10, 0xce, 0x9f, 0xe7, 0xbd, 0x59, 0xa8,
		0x65, 0x00, 0xbe, 0xb5, 0x24, 0xa7, 0x89, 0xae,
	}
	pk := sk.PublicKey()
	pubBytes := pk.Serialize()
	if !bytes.Equal(pubBytes, pkExpected) {
		t.Errorf("got %v, expected %v", keyBytes, pkExpected)
	}

	expectedSigBytes := []byte{
		0x19, 0xc6, 0xad, 0x8b, 0xdd, 0x15, 0x8d, 0x17,
		0xbc, 0x65, 0x0d, 0xc9, 0x82, 0x7a, 0x8a, 0x25,
		0x2c, 0xa1, 0xf9, 0xda, 0x2c, 0x9d, 0xda, 0x20,
		0x58, 0xf9, 0x79, 0x88, 0x09, 0x8e, 0xe3, 0x1d,
		0x68, 0x72, 0xa8, 0x4a, 0x54, 0xf6, 0x89, 0xbc,
		0xcd, 0x64, 0x64, 0xf2, 0x88, 0x89, 0xc0, 0x5e,
		0x07, 0x73, 0x37, 0x32, 0x26, 0x0c, 0x55, 0xde,
		0x64, 0xba, 0x59, 0xe1, 0x04, 0x25, 0x9c, 0x0e,
		0xa4, 0xaa, 0x0f, 0x08, 0x67, 0xfc, 0x6f, 0x8b,
		0x4e, 0x46, 0x85, 0x1a, 0x2c, 0x65, 0x2b, 0x9f,
		0xa1, 0x39, 0x83, 0x23, 0xb4, 0x1a, 0x9c, 0xcf,
		0x23, 0x77, 0xdb, 0x76, 0x41, 0xd2, 0x66, 0x4f,
	}
	msg := []byte{0x07, 0x08, 0x09}
	sig := sk.SignInsecure(msg)
	sigBytes := sig.Serialize()
	if !bytes.Equal(sigBytes, expectedSigBytes) {
		t.Errorf("got %v, expected %v", sigBytes, expectedSigBytes)
	}

	sk2, _ := bls.PrivateKeyFromBytes(result, true)
	keyBytes2 := sk2.Serialize()
	if !bytes.Equal(keyBytes2, result) {
		t.Errorf("got %v, expected %v", keyBytes2, result)
	}

	if !sk.Equal(sk2) {
		t.Error("sk should be equal to sk2")
	}

	sk2.Free()

	sk1, _ := bls.PrivateKeyFromBytes(sk1Bytes, true)
	sk2, _ = bls.PrivateKeyFromBytes(sk2Bytes, true)
	if sk1.Equal(sk2) {
		t.Error("sk1 should NOT be equal to sk2")
	}

	pk1 := sk1.PublicKey()
	pk2 := sk2.PublicKey()

	aggSk, err := bls.PrivateKeyAggregate([]bls.PrivateKey{sk1, sk2}, []bls.PublicKey{pk1, pk2})
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	aggSkBytes := aggSk.Serialize()
	aggSkExpectedBytes := []byte{
		0x66, 0x71, 0xd7, 0x13, 0xd9, 0x34, 0x09, 0x75,
		0xd2, 0x57, 0x77, 0xe4, 0xf7, 0x54, 0x26, 0x6a,
		0xc6, 0x11, 0xb8, 0x3e, 0x90, 0xa6, 0x45, 0x48,
		0xc4, 0x94, 0x7e, 0x57, 0xe0, 0x2c, 0x16, 0x4f,
	}
	if !bytes.Equal(aggSkBytes, aggSkExpectedBytes) {
		t.Errorf("got %v, expected %v", aggSkBytes, aggSkExpectedBytes)
	}

	aggSkIns, err := bls.PrivateKeyAggregateInsecure([]bls.PrivateKey{sk1, sk2})
	if err != nil {
		t.Errorf("got unexpected error: %v", err.Error())
	}
	aggSkInsBytes := aggSkIns.Serialize()
	aggSkInsExpectedBytes := []byte{
		0x52, 0x5c, 0x0a, 0x8d, 0xfe, 0x70, 0x74, 0xd0,
		0xef, 0x8c, 0xc8, 0xed, 0xb3, 0x47, 0xb9, 0xbf,
		0x8f, 0xf6, 0x49, 0xf9, 0xad, 0x89, 0x54, 0x7f,
		0xfd, 0xe1, 0x09, 0xf4, 0x58, 0x35, 0xbd, 0x04,
	}
	if !bytes.Equal(aggSkInsBytes, aggSkInsExpectedBytes) {
		t.Errorf("got %v, expected %v", aggSkInsBytes, aggSkInsExpectedBytes)
	}

	bn1 := new(big.Int).SetBytes([]byte{
		0x2a, 0xc1, 0x24, 0xc0, 0xaa, 0x18, 0x08, 0xe5,
		0x90, 0xff, 0x1f, 0x94, 0xd6, 0x7a, 0x53, 0x97,
		0x0a, 0xe9, 0x82, 0xaa, 0x30, 0xbb, 0xe2, 0x61,
		0xff, 0x1c, 0xb2, 0xad, 0x15, 0xb7, 0x45, 0x2a,
	})
	sk3 := bls.PrivateKeyFromBN(bn1)
	sk3Bytes := sk3.Serialize()
	sk3Expected := []byte{
		0x2a, 0xc1, 0x24, 0xc0, 0xaa, 0x18, 0x08, 0xe5,
		0x90, 0xff, 0x1f, 0x94, 0xd6, 0x7a, 0x53, 0x97,
		0x0a, 0xe9, 0x82, 0xaa, 0x30, 0xbb, 0xe2, 0x61,
		0xff, 0x1c, 0xb2, 0xad, 0x15, 0xb7, 0x45, 0x2a,
	}
	if !bytes.Equal(sk3Bytes, sk3Expected) {
		t.Errorf("got %v, expected %v", sk3Bytes, sk3Expected)
	}

	// test smaller bignum (ensure padded w/zeroes)
	bn2 := new(big.Int).SetBytes([]byte{
		0x2a, 0xc1, 0x24, 0xc0, 0xaa, 0x18, 0x08, 0xe5,
		0x90, 0xff, 0x1f, 0x94, 0xd6, 0x7a, 0x53, 0x97,
		0x0a, 0xe9, 0x82, 0xaa, 0x30, 0xbb, 0xe2, 0x61,
		0xff, 0x1c, 0xb2, 0xad, 0x15, 0xb7,
	})
	sk4 := bls.PrivateKeyFromBN(bn2)
	sk4Bytes := sk4.Serialize()
	sk4Expected := []byte{
		0x00, 0x00, 0x2a, 0xc1, 0x24, 0xc0, 0xaa, 0x18,
		0x08, 0xe5, 0x90, 0xff, 0x1f, 0x94, 0xd6, 0x7a,
		0x53, 0x97, 0x0a, 0xe9, 0x82, 0xaa, 0x30, 0xbb,
		0xe2, 0x61, 0xff, 0x1c, 0xb2, 0xad, 0x15, 0xb7,
	}
	if !bytes.Equal(sk4Bytes, sk4Expected) {
		t.Errorf("got %v, expected %v", sk4Bytes, sk4Expected)
	}

	aggSkIns.Free()
	aggSk.Free()

	sk4.Free()
	sk3.Free()
	sk2.Free()
	sk1.Free()
	pk.Free()
	sk.Free()
}