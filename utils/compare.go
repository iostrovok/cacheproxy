package utils

import (
	"fmt"
	"github.com/iostrovok/cacheproxy/sqlite"
)

type DiffType string

const (
	Ok       DiffType = "equal"     // records are equal and they are in both files.
	ANoB     DiffType = "A_only"    // Key (args) is found in file A only.
	BNoA     DiffType = "B_only"    // Key (args) is found in file B only.
	Body     DiffType = "body"      // Keys are equal, but bodys aren't (last_date were not compared).
	LastDate DiffType = "last_date" // Keys and body are equal, but last_dates aren't.
)

func (d DiffType) String() string {
	return string(d)
}

type Diff struct {
	Diff    DiffType         `json:"diff"`
	Records []*sqlite.Record `json:"record"`
}

// Compare comperes 2 files by args
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
		keysB[allB[i].Args] = allB[i]
	}

	for i := range allA {
		a := allA[i]

		b, find := keysB[allA[i].Args]
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

		keysBFound[b.Args] = true

		if a.Body.Hash != a.Body.Hash {
			diff.Diff = Body
		} else if a.LastDate != a.LastDate {
			diff.Diff = LastDate
		}

		out = append(out, diff)
	}

	for i := range allB {
		if !keysBFound[allB[i].Args] {
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
