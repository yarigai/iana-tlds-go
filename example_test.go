package ianatlds_test

import (
	"encoding/json"
	"fmt"

	ianatlds "github.com/yarigai/iana-tlds-go"
)

func ExampleIsValidEmail() {
	fmt.Println(ianatlds.IsValidEmail("ana@studio.photography"))
	fmt.Println(ianatlds.IsValidEmail("ana@example.xn--p1ai"))
	fmt.Println(ianatlds.IsValidEmail("ana@example.con"))
	// Output:
	// true
	// true
	// false
}

func ExampleValidateEmail() {
	r := ianatlds.ValidateEmail("user@example.xyzzy")
	if !r.Valid {
		fmt.Println(r.Reason, r.TLD)
	}
	out, _ := json.Marshal(ianatlds.ValidateEmail("not-an-email"))
	fmt.Println(string(out))
	// Output:
	// unknown_tld xyzzy
	// {"valid":false,"email":"not-an-email","tld":null,"reason":"invalid_format"}
}
