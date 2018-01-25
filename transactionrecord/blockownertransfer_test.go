// Copyright (c) 2014-2018 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package transactionrecord_test

import (
	"bytes"
	"encoding/json"
	"github.com/bitmark-inc/bitmarkd/currency"
	"github.com/bitmark-inc/bitmarkd/merkle"
	"github.com/bitmark-inc/bitmarkd/transactionrecord"
	"github.com/bitmark-inc/bitmarkd/util"
	"golang.org/x/crypto/ed25519"
	"reflect"
	"testing"
)

// test the packing/unpacking of base record
//
// ensures that pack->unpack returns the same original value
func TestPackBlockOwnerTransfer(t *testing.T) {

	proofedbyAccount := makeAccount(proofedby.publicKey)

	var link merkle.Digest
	err := merkleDigestFromLE("11162a5683fbf8f73d940ddc71e5cdb426789e6c3e0dbf8573a1990924157cdd", &link)
	if nil != err {
		t.Fatalf("hex to link error: %s", err)
	}

	r := transactionrecord.BlockOwnerTransfer{
		Link:    link,
		Version: 1,
		Payments: currency.Map{
			currency.Bitcoin:  "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn",
			currency.Litecoin: "mmCKZS7toE69QgXNs1JZcjW6LFj8LfUbz6",
		},
		Owner: proofedbyAccount,
	}

	expected := []byte{
		0x07, 0x20, 0x11, 0x16, 0x2a, 0x56, 0x83, 0xfb,
		0xf8, 0xf7, 0x3d, 0x94, 0x0d, 0xdc, 0x71, 0xe5,
		0xcd, 0xb4, 0x26, 0x78, 0x9e, 0x6c, 0x3e, 0x0d,
		0xbf, 0x85, 0x73, 0xa1, 0x99, 0x09, 0x24, 0x15,
		0x7c, 0xdd, 0x01, 0x48, 0x01, 0x22, 0x6d, 0x69,
		0x70, 0x63, 0x42, 0x62, 0x46, 0x67, 0x39, 0x67,
		0x4d, 0x69, 0x43, 0x68, 0x38, 0x31, 0x4b, 0x6a,
		0x38, 0x74, 0x71, 0x71, 0x64, 0x67, 0x6f, 0x5a,
		0x75, 0x62, 0x31, 0x5a, 0x4a, 0x52, 0x66, 0x6e,
		0x02, 0x22, 0x6d, 0x6d, 0x43, 0x4b, 0x5a, 0x53,
		0x37, 0x74, 0x6f, 0x45, 0x36, 0x39, 0x51, 0x67,
		0x58, 0x4e, 0x73, 0x31, 0x4a, 0x5a, 0x63, 0x6a,
		0x57, 0x36, 0x4c, 0x46, 0x6a, 0x38, 0x4c, 0x66,
		0x55, 0x62, 0x7a, 0x36, 0x21, 0x13, 0x55, 0xb2,
		0x98, 0x88, 0x17, 0xf7, 0xea, 0xec, 0x37, 0x74,
		0x1b, 0x82, 0x44, 0x71, 0x63, 0xca, 0xaa, 0x5a,
		0x9d, 0xb2, 0xb6, 0xf0, 0xce, 0x72, 0x26, 0x26,
		0x33, 0x8e, 0x5e, 0x3f, 0xd7, 0xf7,
	}

	expectedTxId := merkle.Digest{
		0xf1, 0x57, 0xbd, 0x6d, 0xe2, 0xcd, 0x63, 0xf9,
		0x51, 0xe0, 0xa0, 0x64, 0x91, 0xf2, 0xff, 0x96,
		0xf3, 0xf8, 0x2a, 0xdd, 0x50, 0x0b, 0x47, 0xdf,
		0x68, 0x5f, 0xb0, 0x5c, 0x6e, 0x6a, 0x0d, 0x8e,
	}

	// manually sign the record and attach signature to "expected"
	signature := ed25519.Sign(proofedby.privateKey, expected)
	r.Signature = signature
	//t.Logf("signature: %#v", r.Signature)
	l := util.ToVarint64(uint64(len(signature)))
	expected = append(expected, l...)
	expected = append(expected, signature...)

	// test the packer
	packed, err := r.Pack(proofedbyAccount)
	if nil != err {
		if nil != packed {
			t.Errorf("partial packed:\n%s", util.FormatBytes("expected", packed))
		}
		t.Fatalf("pack error: %s", err)
	}

	// if either of above fail we will have the message _without_ a signature
	if !bytes.Equal(packed, expected) {
		t.Errorf("pack record: %x  expected: %x", packed, expected)
		t.Errorf("*** GENERATED Packed:\n%s", util.FormatBytes("expected", packed))
		t.Fatal("fatal error")
	}

	// check the record type
	if transactionrecord.BlockOwnerTransferTag != packed.Type() {
		t.Fatalf("pack record type: %x  expected: %x", packed.Type(), transactionrecord.BlockOwnerTransferTag)
	}

	t.Logf("Packed length: %d bytes", len(packed))

	// check txIds
	txId := packed.MakeLink()

	if txId != expectedTxId {
		t.Errorf("pack tx id: %#v  expected: %#v", txId, expectedTxId)
		t.Errorf("*** GENERATED tx id:\n%s", util.FormatBytes("expectedTxId", txId[:]))
	}

	// test the unpacker
	unpacked, n, err := packed.Unpack(true)
	if nil != err {
		t.Fatalf("unpack error: %s", err)
	}

	if len(packed) != n {
		t.Errorf("did not unpack all data: only used: %d of: %d bytes", n, len(packed))
	}

	blockOwnerTransfer, ok := unpacked.(*transactionrecord.BlockOwnerTransfer)
	if !ok {
		t.Fatalf("did not unpack to BlockOwnerTransfer")
	}

	// display a JSON version for information
	item := struct {
		TxId               merkle.Digest
		BlockOwnerTransfer *transactionrecord.BlockOwnerTransfer
	}{
		TxId:               txId,
		BlockOwnerTransfer: blockOwnerTransfer,
	}
	b, err := json.MarshalIndent(item, "", "  ")
	if nil != err {
		t.Fatalf("json error: %s", err)
	}

	t.Logf("BlockOwnerTransfer: JSON: %s", b)

	// check that structure is preserved through Pack/Unpack
	// note reg is a pointer here
	if !reflect.DeepEqual(r, *blockOwnerTransfer) {
		t.Errorf("different, original: %v  recovered: %v", r, *blockOwnerTransfer)
	}
}