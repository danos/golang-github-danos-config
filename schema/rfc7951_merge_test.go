// Copyright (c) 2020-2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/config/schema"
	"github.com/danos/config/testutils"
	"github.com/danos/encoding/rfc7951/data"
	yang "github.com/danos/yang/schema"
)

const rfc7951SchemaTemplate = `
module test-merge {
        namespace "urn:vyatta.com:test:merge";
        prefix test-merge;
        organization "AT&T Inc.";
        contact
		"AT&T
		 Postal: 208 S. Akard Street
		 Dallas, TX 25202
		 Web: www.att.com";
        revision 2020-05-08 {
                description "Test schema for rfc7951 merge";
        }
        %s
}
`

func makeSchema(
	t *testing.T,
	schema string,
) (yang.ModelSet, yang.ModelSet, error) {
	sch := bytes.NewBufferString(fmt.Sprintf(rfc7951SchemaTemplate, schema))

	return testutils.NewModelSetSpec(t).
		SetSchemas(sch.Bytes()).
		GenerateModelSets()
}

func TestRFC7951Merge(t *testing.T) {
	const testSchema = `
container testcontainer {
	config false;
        list testlist {
                key nodetag;
                leaf nodetag {
                        type string;
                }
		container cont {
			leaf testleaf {
				type string;
			}
		}
        }
}
`
	_, msFull, err := makeSchema(t, testSchema)
	if err != nil {
		t.Fatal(err)
	}
	array1 := data.ArrayWith(
		data.ObjectWith(
			data.PairNew("nodetag", "foo"),
			data.PairNew("cont", data.ObjectWith(
				data.PairNew("testleaf", "bar"),
			)),
		),
		data.ObjectWith(
			data.PairNew("nodetag", "bar"),
			data.PairNew("cont", data.ObjectWith(
				data.PairNew("testleaf", "baz"),
			)),
		),
	)
	tree1 := data.TreeFromObject(
		data.ObjectWith(
			data.PairNew("test-merge:testcontainer",
				data.ObjectWith(
					data.PairNew("testlist", array1)))))

	array2 := data.ArrayWith(
		data.ObjectWith(
			data.PairNew("nodetag", "bar"),
			data.PairNew("cont", data.ObjectWith(
				data.PairNew("testleaf", "quux"),
			)),
		),
	)
	tree2 := data.TreeFromObject(
		data.ObjectWith(
			data.PairNew("test-merge:testcontainer",
				data.ObjectWith(
					data.PairNew("testlist", array2)))))
	mrgr := schema.NewRFC7951Merger(msFull, tree1)
	mrgr.Merge(tree2)
	out := mrgr.Tree()
	leaf := out.At("/test-merge:testcontainer/testlist[nodetag='bar']/cont/testleaf")
	if leaf.String() != "quux" {
		t.Fatal("merge failed to update required element")
	}
	leaf = out.At("/test-merge:testcontainer/testlist[nodetag='foo']/cont/testleaf")
	if leaf.String() != "bar" {
		t.Fatal("merge updated incorrect element")
	}
}
