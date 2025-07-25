package mapstructure

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ExampleDecode() {
	type Person struct {
		Name   string
		Age    int
		Emails []string
		Extra  map[string]string
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":   "Mitchell",
		"age":    91,
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "mitchellh",
		},
	}

	var result Person
	err := Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output:
	// mapstructure.Person{Name:"Mitchell", Age:91, Emails:[]string{"one", "two", "three"}, Extra:map[string]string{"twitter":"mitchellh"}}
}

func ExampleDecode_errors() {
	type Person struct {
		Name   string
		Age    int
		Emails []string
		Extra  map[string]string
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":   123,
		"age":    "bad value",
		"emails": []int{1, 2, 3},
	}

	var result Person
	err := Decode(input, &result)
	if err == nil {
		panic("should have an error")
	}

	fmt.Println(err.Error())
	// Output:
	// decoding failed due to the following error(s):
	//
	// 'Name' expected type 'string', got unconvertible type 'int'
	// 'Age' expected type 'int', got unconvertible type 'string'
	// 'Emails[0]' expected type 'string', got unconvertible type 'int'
	// 'Emails[1]' expected type 'string', got unconvertible type 'int'
	// 'Emails[2]' expected type 'string', got unconvertible type 'int'
}

func ExampleDecode_metadata() {
	type Person struct {
		Name string
		Age  int
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":  "Mitchell",
		"age":   91,
		"email": "foo@bar.com",
	}

	// For metadata, we make a more advanced DecoderConfig so we can
	// more finely configure the decoder that is used. In this case, we
	// just tell the decoder we want to track metadata.
	var md Metadata
	var result Person
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		panic(err)
	}

	if err := decoder.Decode(input); err != nil {
		panic(err)
	}

	fmt.Printf("Unused keys: %#v", md.Unused)
	// Output:
	// Unused keys: []string{"email"}
}

func ExampleDecode_weaklyTypedInput() {
	type Person struct {
		Name   string
		Age    int
		Emails []string
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON, generated by a weakly typed language
	// such as PHP.
	input := map[string]interface{}{
		"name":   123,                      // number => string
		"age":    "42",                     // string => number
		"emails": map[string]interface{}{}, // empty map => empty array
	}

	var result Person
	config := &DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		panic(err)
	}

	err = decoder.Decode(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output: mapstructure.Person{Name:"123", Age:42, Emails:[]string{}}
}

func ExampleDecode_tags() {
	// Note that the mapstructure tags defined in the struct type
	// can indicate which fields the values are mapped to.
	type Person struct {
		Name string `mapstructure:"person_name"`
		Age  int    `mapstructure:"person_age"`
	}

	input := map[string]interface{}{
		"person_name": "Mitchell",
		"person_age":  91,
	}

	var result Person
	err := Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output:
	// mapstructure.Person{Name:"Mitchell", Age:91}
}

func ExampleDecode_embeddedStruct() {
	// Squashing multiple embedded structs is allowed using the squash tag.
	// This is demonstrated by creating a composite struct of multiple types
	// and decoding into it. In this case, a person can carry with it both
	// a Family and a Location, as well as their own FirstName.
	type Family struct {
		LastName string
	}
	type Location struct {
		City string
	}
	type Person struct {
		Family    `mapstructure:",squash"`
		Location  `mapstructure:",squash"`
		FirstName string
	}

	input := map[string]interface{}{
		"FirstName": "Mitchell",
		"LastName":  "Hashimoto",
		"City":      "San Francisco",
	}

	var result Person
	err := Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s %s, %s", result.FirstName, result.LastName, result.City)
	// Output:
	// Mitchell Hashimoto, San Francisco
}

func ExampleDecode_remainingData() {
	// Note that the mapstructure tags defined in the struct type
	// can indicate which fields the values are mapped to.
	type Person struct {
		Name  string
		Age   int
		Other map[string]interface{} `mapstructure:",remain"`
	}

	input := map[string]interface{}{
		"name":  "Mitchell",
		"age":   91,
		"email": "mitchell@example.com",
	}

	var result Person
	err := Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output:
	// mapstructure.Person{Name:"Mitchell", Age:91, Other:map[string]interface {}{"email":"mitchell@example.com"}}
}

func ExampleDecode_remainingDataDecodeBackToMapInFlatFormat() {
	// Note that the mapstructure tags defined in the struct type
	// can indicate which fields the values are mapped to.
	type Person struct {
		Name  string
		Age   int
		Other map[string]interface{} `mapstructure:",remain"`
	}

	input := map[string]interface{}{
		"name": "Luffy",
		"age":  19,
		"powers": []string{
			"Rubber Man",
			"Conqueror Haki",
		},
	}

	var person Person
	err := Decode(input, &person)
	if err != nil {
		panic(err)
	}

	result := make(map[string]interface{})
	err = Decode(&person, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output:
	// map[string]interface {}{"Age":19, "Name":"Luffy", "powers":[]string{"Rubber Man", "Conqueror Haki"}}
}

func ExampleDecode_omitempty() {
	// Add omitempty annotation to avoid map keys for empty values
	type Family struct {
		LastName string
	}
	type Location struct {
		City string
	}
	type Person struct {
		*Family   `mapstructure:",omitempty"`
		*Location `mapstructure:",omitempty"`
		Age       int
		FirstName string
	}

	result := &map[string]interface{}{}
	input := Person{FirstName: "Somebody"}
	err := Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", result)
	// Output:
	// &map[Age:0 FirstName:Somebody]
}

func ExampleDecode_decodeHookFunc() {
	type PersonLocation struct {
		Latitude   float64
		Longtitude float64
	}

	type Person struct {
		Name     string
		Location PersonLocation
	}

	// Example of parsing messy input: here we have latitude, longitude squashed into
	// a single string field. We write a custom DecodeHookFunc to parse the '#' separated
	// values into a PersonLocation struct.
	input := map[string]interface{}{
		"name":     "Mitchell",
		"location": "-35.2809#149.1300",
	}

	toPersonLocationHookFunc := func() DecodeHookFunc {
		return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
			if t != reflect.TypeOf(PersonLocation{}) {
				return data, nil
			}

			switch f.Kind() {
			case reflect.String:
				xs := strings.Split(data.(string), "#")

				if len(xs) == 2 {
					lat, errLat := strconv.ParseFloat(xs[0], 64)
					lon, errLon := strconv.ParseFloat(xs[1], 64)

					if errLat == nil && errLon == nil {
						return PersonLocation{Latitude: lat, Longtitude: lon}, nil
					}
				} else {
					return data, nil
				}
			}
			return data, nil
		}
	}

	var result Person

	decoder, errDecoder := NewDecoder(&DecoderConfig{
		Metadata:   nil,
		DecodeHook: toPersonLocationHookFunc(), // Here, use ComposeDecodeHookFunc to run multiple hooks.
		Result:     &result,
	})
	if errDecoder != nil {
		panic(errDecoder)
	}

	err := decoder.Decode(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
	// Output:
	// mapstructure.Person{Name:"Mitchell", Location:mapstructure.PersonLocation{Latitude:-35.2809, Longtitude:149.13}}
}
