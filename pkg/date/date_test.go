package date

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateYearLayout(t *testing.T) {
	var testCases = []struct {
		description  string
		layout       string
		isSuccessful bool
	}{
		{
			description:  "should return no error because the layout '06' is valid",
			layout:       "06",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout '2006' is valid",
			layout:       "2006",
			isSuccessful: true,
		},
		{
			description:  "should return an error because the layout '2066' is invalid",
			layout:       "2066",
			isSuccessful: false,
		},
	}

	for _, testCase := range testCases {
		err := validateYearLayout(testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestValidateMonthLayout(t *testing.T) {
	var testCases = []struct {
		description  string
		layout       string
		isSuccessful bool
	}{
		{
			description:  "should return no error because the layout 'Jan' is valid",
			layout:       "Jan",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout 'January' is valid",
			layout:       "January",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout '1' is valid",
			layout:       "1",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout '01' is valid",
			layout:       "01",
			isSuccessful: true,
		},
		{
			description:  "should return an error because the layout '011' is invalid",
			layout:       "011",
			isSuccessful: false,
		},
	}

	for _, testCase := range testCases {
		err := validateMonthLayout(testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestValidateDayOfTheMonthLayout(t *testing.T) {
	var testCases = []struct {
		description  string
		layout       string
		isSuccessful bool
	}{
		{
			description:  "should return no error because the layout '2' is valid",
			layout:       "2",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout '_2' is valid",
			layout:       "_2",
			isSuccessful: true,
		},
		{
			description:  "should return no error because the layout '02' is valid",
			layout:       "02",
			isSuccessful: true,
		},
		{
			description:  "should return an error because the layout '022' is invalid",
			layout:       "022",
			isSuccessful: false,
		},
	}

	for _, testCase := range testCases {
		err := validateDayOfTheMonthLayout(testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestParseDateComponent(t *testing.T) {
	var testCases = []struct {
		description   string
		layout        string
		dateComponent string
		isSuccessful  bool
	}{
		{
			description:   "should return no error because the layout and dateComponent are both correct",
			layout:        "2006",
			dateComponent: "2024",
			isSuccessful:  true,
		},
		{
			description:   "should return an error because the layout is incorrect",
			layout:        "2066",
			dateComponent: "2024",
			isSuccessful:  false,
		},
		{
			description:   "should return an error because the dateComponent is incorrect",
			layout:        "2006",
			dateComponent: "1",
			isSuccessful:  false,
		},
	}

	for _, testCase := range testCases {
		_, err := ParseDateComponent(testCase.layout, testCase.dateComponent)

		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestParseYear(t *testing.T) {
	var testCases = []struct {
		description  string
		year         string
		layout       string
		isSuccessful bool
		outputValue  uint16
	}{
		{
			description:  "should return '0' because layout is incorrect",
			year:         "2024",
			layout:       "2066",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should return '0' because year is incorrect",
			year:         "1",
			layout:       "2006",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should succeed and return the correct year",
			year:         "2024",
			layout:       "2006",
			isSuccessful: true,
			outputValue:  2024,
		},
	}

	for _, testCase := range testCases {
		y, err := ParseYear(testCase.year, testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, y, testCase.outputValue)

			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestParseMonth(t *testing.T) {
	var testCases = []struct {
		description  string
		month        string
		layout       string
		isSuccessful bool
		outputValue  uint8
	}{
		{
			description:  "should return '0' because layout is incorrect",
			month:        "1",
			layout:       "2066",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should return '0' because month is incorrect",
			month:        "13",
			layout:       "1",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should succeed and return the correct month",
			month:        "5",
			layout:       "1",
			isSuccessful: true,
			outputValue:  5,
		},
	}

	for _, testCase := range testCases {
		m, err := ParseMonth(testCase.month, testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, m, testCase.outputValue)

			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestParseDayOfTheMonth(t *testing.T) {
	var testCases = []struct {
		description  string
		day          string
		layout       string
		isSuccessful bool
		outputValue  uint8
	}{
		{
			description:  "should return '0' because layout is incorrect",
			day:          "1",
			layout:       "2066",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should return '0' because day is incorrect",
			day:          "32",
			layout:       "1",
			isSuccessful: false,
			outputValue:  0,
		},
		{
			description:  "should succeed and return the correct day",
			day:          "5",
			layout:       "2",
			isSuccessful: true,
			outputValue:  5,
		},
	}

	for _, testCase := range testCases {
		d, err := ParseDayOfTheMonth(testCase.day, testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, d, testCase.outputValue)

			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func TestIsLeapYear(t *testing.T) {
	var testCases = []struct {
		description  string
		year         string
		layout       string
		isSuccessful bool
		isTrue       bool
	}{
		{
			description:  "should return 'false' because layout is incorrect",
			year:         "2024",
			layout:       "2066",
			isSuccessful: false,
			isTrue:       false,
		},
		{
			description:  "should return 'false' because year is incorrect",
			year:         "1",
			layout:       "2006",
			isSuccessful: false,
			isTrue:       false,
		},
		{
			description:  "should return 'false' because year can not be divided by 400 or 4",
			year:         "2023",
			layout:       "2006",
			isSuccessful: true,
			isTrue:       false,
		},
		{
			description:  "should return 'false' because year can be divided by 100",
			year:         "2100",
			layout:       "2006",
			isSuccessful: true,
			isTrue:       false,
		},
		{
			description:  "should return 'true' because year can be divided by 400",
			year:         "2000",
			layout:       "2006",
			isSuccessful: true,
			isTrue:       true,
		},
		{
			description:  "should return 'true' because year can be divided by 4 and not by 100",
			year:         "2024",
			layout:       "2006",
			isSuccessful: true,
			isTrue:       true,
		},
	}

	for _, testCase := range testCases {
		condition, err := IsLeapYear(testCase.year, testCase.layout)

		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, condition, testCase.isTrue)

			if testCase.isSuccessful {
				assert.NoError(t, err, "Expected no error, but got an error")
			} else {
				assert.Error(t, err, "Expected an error, but got no error")
			}
		})
	}
}

func Example_parseDateComponent() {
	/*
		ParseDateComponent parses year, month or day strings and returns a
		time object.

		- #### Parameters

		  - layout (string): The layout specifying the format of the date component.
		For example, "2006" or "06".
		  - date component (string): The date component to be parsed. For example,
		"2024".

		- #### Returns

		  - time.Time: A date object representing the parsed date component.
		  - error: An error, if any, encountered during the parsing of the
		date component string.
	*/

	date, err := ParseDateComponent("2006", "2024")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(date) // Output: 2024-01-01 00:00:00 +0000 UTC
}

func Example_parseYear() {
	/*
		ParseYear parses year string

		- #### Parameters

		  - year (string): The year string to be parsed.
		  - layout (string): The layout specifying the format of the year string. For example, "2006" or "06".

		- #### Returns

		  - year (uint16): A uint16 number representing the parsed year.
		  - error: An error, if any, encountered during the parsing of the year
		string.
	*/

	year, err := ParseYear("2024", "2006")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(year) // Output: 2024
}

func Example_parseMonth() {
	/*
		ParseMonth parses month string

		- #### Parameters

		  - month (string): The month string to be parsed.
		  - layout (string): The layout specifying the format of the month string. For example, "Jan", "January", "01" or "1"

		- #### Returns

		  - month (uint8): A uint8 number representing the parsed month.
		  - error: An error, if any, encountered during the parsing of the month
		string.
	*/

	month, err := ParseMonth("5", "1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(month) // Output: 5
}

func Example_parseDayOfTheMonth() {
	/*
		ParseDayOfTheMonth parses day string

		- #### Parameters

		  - day (string): The day string to be parsed.
		  - layout (string): The layout specifying the format of the day string. For example, "2", "_2" or "02"

		- #### Returns

		  - day (uint8): A uint8 number representing the parsed day.
		  - error: An error, if any, encountered during the parsing of the day string.
	*/

	day, err := ParseDayOfTheMonth("15", "2")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(day) // Output: 15
}

func Example_isLeapYear() {
	/*
			IsLeapYear determines whether a specific year is a leap year.

		- #### Parameters

		  - year (string): The year to be checked for leap year status.
		  - layout (string): The layout specifying the format of the year string. For example, "2006" or "06".

		- #### Returns

		  - bool: A boolean indicating whether the specified year is a leap year (true) or not (false).
		  - error: An error, if any, encountered during the parsing of the year string.
	*/

	leap, err := IsLeapYear("2024", "2006")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(leap) // Output: true
}
