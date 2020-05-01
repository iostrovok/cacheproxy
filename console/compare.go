package main

import (
	"fmt"
	"log"

	"github.com/iostrovok/cacheproxy/utils"
)

/*
	Show deference requests between 2 files.
*/

func main() {

	fileA := "./test1.db"
	fileB := "./test1.db"

	res, err := utils.Compare(fileA, fileB)

	if err != nil {
		log.Fatal(err)
	}

	for i, r := range res {
		rec := r.Records[0]
		fmt.Printf("%d] %s: %s\n", i, r.Diff.String(), rec.ID)
		if r.Diff != utils.Ok {
			fmt.Printf("\n\n%s\n\n", string(rec.Body.Request))
		}
	}
}
