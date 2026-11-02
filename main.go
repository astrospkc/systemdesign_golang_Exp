package main

import (
	"fmt"
	goconceptexprace "go_exp/experiments/go_concept_exp/race"
)

func main() {

	fmt.Println("Hello, Go project initialized!")
	// run cancellation lab
	// observablity.Observe()
	// goconceptexp.RatelimitHttpExp()
	// goconceptexp.Performance()
	goconceptexprace.Race()

}
