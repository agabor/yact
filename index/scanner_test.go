package index

import (
	"reflect"
	"testing"
)

const scannerGoSource = `package sample

type Box struct {
	Value int
}

func (b *Box) Get() int {
	return b.Value
}

func helper() int {
	return 1
}

func Make() *Box {
	return &Box{}
}
`

const scannerJSSource = `class Store {
  load() {
    return 1;
  }

  #secret() {
    return 2;
  }
}
`

const scannerJSPrivateOnlySource = `class Store {
  #secret() {
    return 2;
  }
}
`

const scannerPHPSource = `<?php
class Store {
    public function load() {
        return 1;
    }
    private function reset() {
        return 0;
    }
}
`

func TestScanSignaturesGoTopLevelFunctions(t *testing.T) {
	want := []IndexSignature{
		{Text: "func (b *Box) Get() int", Depth: 0},
		{Text: "func helper() int", Depth: 0},
		{Text: "func Make() *Box", Depth: 0},
	}
	got := scanSignatures(scannerGoSource, goSignatureRules)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanSignatures() = %+v, want %+v", got, want)
	}
}

func TestScanSignaturesGoPublicOnly(t *testing.T) {
	rules := goSignatureRules
	rules.PublicOnly = true
	want := []IndexSignature{
		{Text: "func (b *Box) Get() int", Depth: 0},
		{Text: "func Make() *Box", Depth: 0},
	}
	got := scanSignatures(scannerGoSource, rules)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanSignatures() = %+v, want %+v", got, want)
	}
}

func TestScanSignaturesJSNestingDepth(t *testing.T) {
	tests := []struct {
		name       string
		publicOnly bool
		want       []IndexSignature
	}{
		{
			name:       "private method is listed when not public only",
			publicOnly: false,
			want: []IndexSignature{
				{Text: "class Store", Depth: 0},
				{Text: "load()", Depth: 1},
				{Text: "#secret()", Depth: 1},
			},
		},
		{
			name:       "private method is omitted when public only",
			publicOnly: true,
			want: []IndexSignature{
				{Text: "class Store", Depth: 0},
				{Text: "load()", Depth: 1},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := jsSignatureRules
			rules.PublicOnly = tc.publicOnly
			got := scanSignatures(scannerJSSource, rules)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("scanSignatures() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestScanSignaturesJSClassWithOnlyPrivateMembersIsOmitted(t *testing.T) {
	rules := jsSignatureRules
	rules.PublicOnly = true
	got := scanSignatures(scannerJSPrivateOnlySource, rules)
	if len(got) != 0 {
		t.Errorf("scanSignatures() = %+v, want no signatures", got)
	}
}

func TestScanSignaturesPHPVisibility(t *testing.T) {
	tests := []struct {
		name       string
		publicOnly bool
		want       []IndexSignature
	}{
		{
			name:       "all members are listed without visibility filter",
			publicOnly: false,
			want: []IndexSignature{
				{Text: "class Store", Depth: 0},
				{Text: "public function load()", Depth: 1},
				{Text: "private function reset()", Depth: 1},
			},
		},
		{
			name:       "private member is omitted when public only",
			publicOnly: true,
			want: []IndexSignature{
				{Text: "class Store", Depth: 0},
				{Text: "public function load()", Depth: 1},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := phpSignatureRules
			rules.PublicOnly = tc.publicOnly
			got := scanSignatures(scannerPHPSource, rules)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("scanSignatures() = %+v, want %+v", got, tc.want)
			}
		})
	}
}