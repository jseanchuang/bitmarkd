// Copyright (c) 2014-2018 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpccalls

import (
	"github.com/bitmark-inc/bitmarkd/merkle"
	"github.com/bitmark-inc/bitmarkd/rpc"
)

type ProvenanceData struct {
	TxId       string
	Count      int
	Identities map[string]string
}

type ProvenanceReply struct {
	Data []provenanceItem `json:"data"`
}

type provenanceItem struct {
	rpc.ProvenanceRecord
	Identity string `json:"_IDENTITY"`
}

func (client *Client) GetProvenance(provenanceConfig *ProvenanceData) (*ProvenanceReply, error) {

	var txId merkle.Digest
	if err := txId.UnmarshalText([]byte(provenanceConfig.TxId)); nil != err {
		return nil, err
	}

	provenanceArgs := rpc.ProvenanceArguments{
		TxId:  txId,
		Count: provenanceConfig.Count,
	}

	client.printJson("Provenance Request", provenanceArgs)

	var reply rpc.ProvenanceReply
	err := client.client.Call("Bitmark.Provenance", provenanceArgs, &reply)
	if nil != err {
		return nil, err
	}

	client.printJson("Provenance Reply", reply)

	r := &ProvenanceReply{
		Data: make([]provenanceItem, len(reply.Data)),
	}

	for i, d := range reply.Data {
		r.Data[i].ProvenanceRecord = d

		m := d.Data.(map[string]interface{})
		owner := m["owner"]
		if nil == owner {
			owner = m["registrant"]
		}
		if s, ok := owner.(string); ok {
			r.Data[i].Identity = provenanceConfig.Identities[s]
		}
	}

	return r, nil
}
