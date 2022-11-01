package main

import (
	"fmt"
	"sort"

	"github.com/guardian/simple-interactive-deliverables/common"
)

func main() {
	numericIdList := make([]int32, 100)
	for i := 0; i < 100; i++ {
		//println(common.GenerateStringId())
		//fmt.Printf("%d\n", common.GenerateNumericId())
		numericIdList[i] = common.GenerateNumericId()
	}

	sort.Slice(numericIdList, func(a, b int) bool {
		return numericIdList[a] < numericIdList[b]
	})

	for _, v := range numericIdList {
		fmt.Printf("%d\n", v)
	}
}
