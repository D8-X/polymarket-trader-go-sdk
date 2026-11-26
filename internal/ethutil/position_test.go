package ethutil

import (
	"encoding/hex"
	"testing"
)

func TestIsV2PositionID(t *testing.T) {
	cases := map[string]bool{
		"714841671368790317633885189985615663717485908384619261899520210089206087680":   true,
		"714841671368790317633885189985615663717485908384619261899520210089206087681":   true,
		"32338220190071351435772801779725302244575775216413325951443816017994629993401": false,
		"25659310674993675562345759665114759892400026242514633218387667107987341231962": false,
		"not-a-number": false,
		"-1":           false,
		"":             false,
	}
	for id, want := range cases {
		if got := IsV2PositionID(id); got != want {
			t.Errorf("IsV2PositionID(%q)=%v want %v", id, got, want)
		}
	}
}

func TestV2ConditionWord(t *testing.T) {
	const want = "019496073df9798fff1e7dafc6a51806d9000000000000000000000000000000"
	for _, in := range []string{
		"0x019496073df9798fff1e7dafc6a51806d9000000000000000000000000000000",
		"0x019496073df9798fff1e7dafc6a51806d9000000000000000000000000000001",
		"0x019496073df9798fff1e7dafc6a51806d90000000000000000000000000000",
	} {
		got, err := V2ConditionWord(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if hex.EncodeToString(got) != want {
			t.Errorf("%s: got %x want %s", in, got, want)
		}
	}
	if _, err := V2ConditionWord("0xa467b14d51f01b957109d9cbb1d6c124fab2a089d52ed8f471d23c2812e743b7"); err == nil {
		t.Error("expected an error for a CTF condition id")
	}
}

func TestV2PositionID(t *testing.T) {
	word, err := V2ConditionWord("0x019496073df9798fff1e7dafc6a51806d9000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if got := V2PositionID(word, 0); got != "714841671368790317633885189985615663717485908384619261899520210089206087680" {
		t.Errorf("yes: got %s", got)
	}
	if got := V2PositionID(word, 1); got != "714841671368790317633885189985615663717485908384619261899520210089206087681" {
		t.Errorf("no: got %s", got)
	}
}
