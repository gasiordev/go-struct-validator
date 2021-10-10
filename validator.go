package validator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type FieldValidation struct {
	lenMin int
	lenMax int
	valMin int64
	valMax int64
	regexp *regexp.Regexp
	flags  int64
}

// values used with flags
const ValMinNotNil = 2
const ValMaxNotNil = 4
const Required = 8
const Email = 16

// values for invalid field flags
const FailLenMin = 2
const FailLenMax = 4
const FailValMin = 8
const FailValMax = 16
const FailEmpty = 32
const FailRegexp = 64
const FailEmail = 128
const FailZero = 256

type ValidationOptions struct {
	RestrictFields     map[string]bool
	OverwriteFieldTags map[string]map[string]string
	OverwriteTagName   string
	ValidateWhenSuffix bool
}

func Validate(obj interface{}, options *ValidationOptions) (bool, map[string]int) {
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()

	tagName := "validation"
	if options != nil && options.OverwriteTagName != "" {
		tagName = options.OverwriteTagName
	}

	invalidFields := map[string]int{}
	valid := true

	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)
		fieldKind := field.Type.Kind()

		// check if only specified field should be checked
		if options != nil && len(options.RestrictFields) > 0 && !options.RestrictFields[field.Name] {
			continue
		}
		// validate only ints and string
		if !isNotInt(fieldKind) && !isNotString(fieldKind) {
			continue
		}

		validation := FieldValidation{}
		validation.lenMin = -1
		validation.lenMax = -1

		// get tag values
		tagVal := field.Tag.Get(tagName)
		tagRegexpVal := field.Tag.Get(tagName + "_regexp")
		if options != nil && len(options.OverwriteFieldTags) > 0 {
			if len(options.OverwriteFieldTags[field.Name]) > 0 {
				if options.OverwriteFieldTags[field.Name][tagName] != "" {
					tagVal = options.OverwriteFieldTags[field.Name][tagName]
				}
				if options.OverwriteFieldTags[field.Name][tagName+"_regexp"] != "" {
					tagRegexpVal = options.OverwriteFieldTags[field.Name][tagName+"_regexp"]
				}
			}
		}

		setValidationFromTag(&validation, tagVal)
		if tagRegexpVal != "" {
			validation.regexp = regexp.MustCompile(tagRegexpVal)
		}

		if options != nil && options.ValidateWhenSuffix {
			if strings.HasSuffix(field.Name, "Email") {
				validation.flags = validation.flags | Email
			}
			if strings.HasSuffix(field.Name, "Price") && validation.valMin == 0 && validation.valMax == 0 && validation.flags&ValMinNotNil == 0 && validation.flags&ValMaxNotNil == 0 {
				validation.valMin = 0
				validation.flags = validation.flags | ValMinNotNil
			}
		}

		fieldValid, failureFlags := validateValue(v.Elem().FieldByName(field.Name), &validation)
		if !fieldValid {
			valid = false
			invalidFields[field.Name] = failureFlags
		}
	}

	return valid, invalidFields
}

func validateValue(value reflect.Value, validation *FieldValidation) (bool, int) {
	minCanBeZero := false
	maxCanBeZero := false
	if validation.flags&ValMinNotNil > 0 {
		minCanBeZero = true
	}
	if validation.flags&ValMaxNotNil > 0 {
		maxCanBeZero = true
	}

	if validation.flags&Required > 0 {
		if value.Type().Name() == "string" && value.String() == "" {
			return false, FailEmpty
		}
		if strings.HasPrefix(value.Type().Name(), "int") && value.Int() == 0 && !minCanBeZero && !maxCanBeZero && validation.valMin == 0 && validation.valMax == 0 {
			return false, FailZero
		}
	}

	if value.Type().Name() == "string" {
		if validation.lenMin > 0 && len(value.String()) < validation.lenMin {
			return false, FailLenMin
		}
		if validation.lenMax > 0 && len(value.String()) > validation.lenMax {
			return false, FailLenMax
		}

		if validation.regexp != nil {
			if !validation.regexp.MatchString(value.String()) {
				return false, FailRegexp
			}
		}

		if validation.flags&Email > 0 {
			var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
			if !emailRegex.MatchString(value.String()) {
				return false, FailEmail
			}
		}
	}

	if strings.HasPrefix(value.Type().Name(), "int") {
		if (validation.valMin != 0 || minCanBeZero) && validation.valMin > value.Int() {
			return false, FailValMin
		}
		if (validation.valMax != 0 || maxCanBeZero) && validation.valMax < value.Int() {
			return false, FailValMax
		}
	}

	return true, 0
}

func setValidationFromTag(v *FieldValidation, tag string) {
	opts := strings.SplitN(tag, " ", -1)
	for _, opt := range opts {
		if opt == "req" {
			v.flags = v.flags | Required
		}
		if opt == "email" {
			v.flags = v.flags | Email
		}
		for _, valOpt := range []string{"lenmin", "lenmax", "valmin", "valmax", "regexp"} {
			if strings.HasPrefix(opt, valOpt+":") {
				val := strings.Replace(opt, valOpt+":", "", 1)
				if valOpt == "regexp" {
					v.regexp = regexp.MustCompile(val)
					continue
				}

				i, err := strconv.Atoi(val)
				if err != nil {
					continue
				}
				switch valOpt {
				case "lenmin":
					v.lenMin = i
				case "lenmax":
					v.lenMax = i
				case "valmin":
					v.valMin = int64(i)
					if i == 0 {
						v.flags = v.flags | ValMinNotNil
					}
				case "valmax":
					v.valMax = int64(i)
					if i == 0 {
						v.flags = v.flags | ValMaxNotNil
					}
				}
			}
		}
	}
}

func isNotInt(k reflect.Kind) bool {
	if k == reflect.Int64 || k == reflect.Int32 || k == reflect.Int16 || k == reflect.Int8 || k == reflect.Int || k == reflect.Uint64 || k == reflect.Uint32 || k == reflect.Uint16 || k == reflect.Uint8 || k == reflect.Uint {
		return true
	}
	return false
}

func isNotString(k reflect.Kind) bool {
	if k == reflect.String {
		return true
	}
	return false
}
