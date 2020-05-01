package utils

import (
	"fmt"

	"github.com/iostrovok/cacheproxy/sqlite"
)

type DiffType string

const (
	Ok   DiffType = "equal"  // records are equal and they are in both files.
	ANoB DiffType = "A_only" // Key is found in file A only.
	BNoA DiffType = "B_only" // Key is found in file B only.
	Body DiffType = "body"   // Keys are equal, but bodies aren't (last_date were not compared).
)

func (d DiffType) String() string {
	return string(d)
}

type Diff struct {
	Diff    DiffType         `json:"diff"`
	Records []*sqlite.Record `json:"record"`
}

// Compare comperes 2 files by id and body
func Compare(pathA, pathB string) ([]*Diff, error) {

	out := make([]*Diff, 0)

	connA, err := sqlite.Conn(pathA)
	if err != nil {
		return nil, err
	}

	connB, err := sqlite.Conn(pathB)
	if err != nil {
		return nil, err
	}

	allA, err := connA.SelectAll()
	allB, err := connB.SelectAll()

	keysBFound := map[string]bool{}
	keysB := map[string]*sqlite.Record{}
	for i := range allB {
		keysB[allB[i].ID] = allB[i]
	}

	for i := range allA {
		a := allA[i]

		b, find := keysB[allA[i].ID]
		if !find {
			diff := &Diff{
				Diff:    ANoB,
				Records: []*sqlite.Record{a},
			}
			out = append(out, diff)
			continue
		}

		diff := &Diff{
			Diff:    Ok,
			Records: []*sqlite.Record{a, b},
		}

		keysBFound[b.ID] = true

		if a.Body.Hash != a.Body.Hash {
			diff.Diff = Body
		}
		out = append(out, diff)
	}

	for i := range allB {
		if !keysBFound[allB[i].ID] {
			diff := &Diff{
				Diff:    BNoA,
				Records: []*sqlite.Record{allB[i]},
			}
			out = append(out, diff)
		}
	}

	fmt.Printf("allA: %d\n", len(allA))
	fmt.Printf("allB: %d\n", len(allB))

	return out, err
}
