package validator

import (
	"testing"
)

type Test1 struct {
	FirstName     string `validation:"req lenmin:5 lenmax:25"`
	LastName      string `validation:"req lenmin:2 lenmax:50"`
	Age           int    `validation:"req valmin:18 valmax:150"`
	Price         int    `validation:"req valmin:0 valmax:9999"`
	PostCode      string `validation:"req" validation_regexp:"^[0-9][0-9]-[0-9][0-9][0-9]$"`
	Email         string `validation:"req email"`
	BelowZero     int    `validation:"valmin:-6 valmax:-2"`
	DiscountPrice int    `validation:"valmin:0 valmax:8000"`
	Country       string `validation_regexp:"^[A-Z][A-Z]$"`
	County        string `validation:"lenmax:40"`
}

type Test2 struct {
	FirstName     string `mytag:"req lenmin:5 lenmax:25"`
	LastName      string `mytag:"req lenmin:2 lenmax:50"`
	Age           int    `mytag:"req valmin:18 valmax:150"`
	Price         int    `mytag:"req valmin:0 valmax:9999"`
	PostCode      string `mytag:"req" mytag_regexp:"^[0-9][0-9]-[0-9][0-9][0-9]$"`
	Email         string `mytag:"req email"`
	BelowZero     int    `mytag:"valmin:-6 valmax:-2"`
	DiscountPrice int    `mytag:"valmin:0 valmax:8000"`
	Country       string `mytag_regexp:"^[A-Z][A-Z]$"`
	County        string `mytag:"lenmax:40"`
}

func TestWithDefaultValues(t *testing.T) {
	s := Test1{}
	expectedBool := false
	expectedFailedFields := map[string]int{
		"FirstName": FailEmpty,
		"LastName": FailEmpty,
		"Age": FailZero,
		"PostCode": FailEmpty,
		"Email": FailEmpty,
		"Country": FailRegexp,
		"BelowZero": FailValMax,
	}
	compare(&s, expectedBool, expectedFailedFields, nil, nil, "", t)
}

func TestWithInvalidValues(t *testing.T) {
	s := Test1{
		FirstName: "123456789012345678901234567890",
		LastName: "b",
		Age: 15,
		Price: 0,
		PostCode: "AA123",
		Email: "invalidEmail",
		BelowZero: 8,
		DiscountPrice: 9999,
		Country: "Tokelau",
		County: "",
	}
	expectedBool := false
	expectedFailedFields := map[string]int{
		"FirstName": FailLenMax,
		"LastName": FailLenMin,
		"Age": FailValMin,
		"PostCode": FailRegexp,
		"Email": FailEmail,
		"BelowZero": FailValMax,
		"DiscountPrice": FailValMax,
		"Country": FailRegexp,
	}
	compare(&s, expectedBool, expectedFailedFields, nil, nil, "", t)
}

func TestWithValidValues(t *testing.T) {
	s := Test1{
		FirstName: "Johnny",
		LastName: "Smith",
		Age: 35,
		Price: 0,
		PostCode: "43-155",
		Email: "john@example.com",
		BelowZero: -4,
		DiscountPrice: 8000,
		Country: "GB",
		County: "Enfield",
	}
	expectedBool := true
	expectedFailedFields := map[string]int{}
	compare(&s, expectedBool, expectedFailedFields, nil, nil, "", t)
}

func TestWithInvalidValuesAndFieldRestriction(t *testing.T) {
	s := Test1{
		FirstName: "123456789012345678901234567890",
		LastName: "b",
		Age: 15,
		Price: 0,
		PostCode: "AA123",
		Email: "invalidEmail",
		BelowZero: 8,
		DiscountPrice: 9999,
		Country: "Tokelau",
		County: "",
	}
	expectedBool := false
	expectedFailedFields := map[string]int{
		"FirstName": FailLenMax,
		"LastName": FailLenMin,
	}
	compare(&s, expectedBool, expectedFailedFields, map[string]bool{
		"FirstName": true,
		"LastName": true,
	}, nil, "", t)
}


func TestWithInvalidValuesAndFieldRestrictionAndOverwrittenFieldTags(t *testing.T) {
	s := Test1{
		FirstName: "123456789012345678901234567890",
		LastName: "b",
		Age: 15,
		Price: 0,
		PostCode: "AA123",
		Email: "invalidEmail",
		BelowZero: 8,
		DiscountPrice: 9999,
		Country: "Tokelau",
		County: "",
	}
	expectedBool := false
	expectedFailedFields := map[string]int{
		"LastName": FailLenMin,
	}
	compare(&s, expectedBool, expectedFailedFields, map[string]bool{
		"FirstName": true,
		"LastName": true,
	}, map[string]map[string]string{
		"FirstName": map[string]string{
			"validation": "req lenmin:4 lenmax:100",
		},
	}, "", t)
}

func TestWithInvalidValuesAndOverwrittenTagName(t *testing.T) {
	s := Test2{
		FirstName: "123456789012345678901234567890",
		LastName: "b",
		Age: 15,
		Price: 0,
		PostCode: "AA123",
		Email: "invalidEmail",
		BelowZero: 8,
		DiscountPrice: 9999,
		Country: "Tokelau",
		County: "",
	}
	expectedBool := false
	expectedFailedFields := map[string]int{
		"FirstName": FailLenMax,
		"LastName": FailLenMin,
		"Age": FailValMin,
		"PostCode": FailRegexp,
		"Email": FailEmail,
		"BelowZero": FailValMax,
		"DiscountPrice": FailValMax,
		"Country": FailRegexp,
	}
	compare(&s, expectedBool, expectedFailedFields, nil, nil, "mytag", t)
}

func compare(s interface{}, expectedBool bool, expectedFailedFields map[string]int, restrictFields map[string]bool, overwriteFieldTags map[string]map[string]string, overwriteTagName string, t *testing.T) {
	valid, failedFields := Validate(s, restrictFields, overwriteFieldTags, overwriteTagName)
	if valid != expectedBool {
		t.Fatalf("Validate returned invalid boolean value")
	}
	compareFailedFields(failedFields, expectedFailedFields, t)
}

func compareFailedFields(failedFields map[string]int, expectedFailedFields map[string]int, t *testing.T) {
	if len(failedFields) != len(expectedFailedFields) {
		t.Fatalf("Validate returned invalid number of failed fields %d where it should be %d", len(failedFields), len(expectedFailedFields))
	}
	for k, v := range expectedFailedFields {
		if failedFields[k] != v {
			t.Fatalf("Validate returned invalid failure flag of %d where it should be %d for %s", failedFields[k], v, k)
		}
	}
}