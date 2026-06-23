package aigc_test

import (
	"fmt"

	"github.com/gtkit/aigc"
)

func ExampleIdentifier_Validate() {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "PRODUCER-001"}
	fmt.Println(id.Validate())
	// Output:
	// <nil>
}
