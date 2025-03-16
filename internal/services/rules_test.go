package rules

import (
	"testing"

	"receipt-processor-challenge-jase180/internal/models"
)

func TestPointsForRetailerName(t *testing.T) {
	tests := []struct {
		name     string
		retailer string
		expected int
	}{
		{"All Letters", "abcdefghijklmnopqrstuvwxyz", 26},
		{"All Letters Uppercase", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", 26},
		{"All Numbers", "0123456789", 10},
		{"All Symbols", "!@#$%^&*()", 0},
		{"Letters and Numbers", "abcdefghijklmnopqrstuvwxyz0123456789", 36},
		{"Letters and Symbols", "abcdefghijklmnopqrstuvwxyz!@#$%^&*()", 26},
		{"Letters, Numbers and Symbols mixed with white spaces and dashes", "ab@c!d012#3efghi45jkl*mn67opqrst8u9vw@xyz - -", 36},
		{"Letters(upper and lower), Numbers and Symbols mixed with white spaces and dashes", "aB@c!D012#3eFGhi45jkl*mn67oPQRst8u9vw@xyz - -", 36},
		{"Empty String", "", 0},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := PointsForRetailerName(testCase.retailer)
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestPointsForRoundTotal(t *testing.T) {
	tests := []struct {
		name          string
		total         string
		expected      int
		conversionErr bool // true if we want an error to show up
	}{
		{"Round dollar", "42.00", 50, false},
		{"Not round dollar", "42.01", 0, false},
		{"Floating point check", "9.99999999999999999999999", 0, false}, //fails gracefully from regex
		{"Not a number", "!@#$%^&*()", 0, false},                        //fails gracefully from regex
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := PointsForRoundTotal(testCase.total)
			if (err != nil) != testCase.conversionErr { // if there is an error and we didn't get error
				t.Errorf("unexpected error, error was %v, want %v", err, testCase.conversionErr)
			}
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestPointsForQuarterMultiple(t *testing.T) {
	tests := []struct {
		name          string
		total         string
		expected      int
		conversionErr bool // true if we want an error to show up
	}{
		{"Quarters with round dollar", "42.00", 25, false},
		{"Quarters with not round dollar", "42.25", 25, false},
		{"Not round dollar", "42.42", 0, false},
		{"Floating point error check", "42.2499999999999999999999999999", 0, false},
		{"Not a number", "!@#$%^&*()", 0, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := PointsForQuarterMultiple(testCase.total)
			if (err != nil) != testCase.conversionErr { // if there is an error and we didn't get error
				t.Errorf("unexpected error, error was %v, want %v", err, testCase.conversionErr)
			}
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestPointsForEveryTwoItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []models.Item
		expected int
	}{
		{"5 items", []models.Item{{}, {}, {}, {}, {}}, 10},
		{"4 items", []models.Item{{}, {}, {}, {}}, 10},
		{"0 items", []models.Item{}, 0},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := PointsForEveryTwoItems(testCase.items)
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestPointsForItemDescription(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		price         string
		expected      int
		conversionErr bool // true if we want an error to show up
	}{
		{"Length multiple of 3", "abc", "3.00", 1, false}, // 3 * 0.2 = 0.6 round up to 1
		{"Length is not multiple of 3", "abcde", "3.00", 0, false},
		{"Trimmed length is multiple of 3", "  abc  ", "6.42", 2, false}, // 6.42 * 0.2 = 1.284 round up to 2
		{"not a number", "abc", "!@#$%^", 0, true},
		{"Empty description", "", "10.00", 0, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// write test case into the item for function's argument
			item := models.Item{
				ShortDescription: testCase.description,
				Price:            testCase.price,
			}

			result, err := PointsForItemDescription(item)

			if (err != nil) != testCase.conversionErr {
				t.Errorf("unexpected error, got %v, want %v", err, testCase.conversionErr) // check conversion error
			}

			if result != testCase.expected {
				t.Errorf("PointsForItemDescription(%q, %q) = %v; want %v",
					testCase.description, testCase.price, result, testCase.expected)
			}
		})
	}
}

func TestPointsForOddDay(t *testing.T) {
	tests := []struct {
		name          string
		date          string
		expected      int
		conversionErr bool // true if we want an error to show up
	}{
		{"Odd date", "2025-03-15", 6, false},
		{"Even date", "2025-03-16", 0, false},
		{"Leap year", "2024-02-29", 6, false},
		{"Invalid date", "42.42", 0, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := PointsForOddDay(testCase.date)
			if (err != nil) != testCase.conversionErr { // if there is an error and we didn't get error
				t.Errorf("unexpected error, error was %v, want %v", err, testCase.conversionErr)
			}
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestPointsForTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		time          string
		expected      int
		conversionErr bool // true if we want an error to show up
	}{
		{"Within range", "14:01", 10, false},
		{"Outside range", "14:00", 0, false},
		{"Edge time within", "15:59", 10, false},
		{"Edge time outside", "16:00", 0, false},
		{"Invalid time", "42.42", 0, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := PointsForTimeRange(testCase.time)
			if (err != nil) != testCase.conversionErr { // if there is an error and we didn't get error
				t.Errorf("unexpected error, error was %v, want %v", err, testCase.conversionErr)
			}
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		name     string
		receipt  models.Receipt
		expected int
	}{
		{
			name: "Target Receipt from README",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
				Total: "35.35",
			},
			expected: 28,
		},
		{
			name: "M&M Corner Market Receipt from README",
			receipt: models.Receipt{
				Retailer:     "M&M Corner Market",
				PurchaseDate: "2022-03-20",
				PurchaseTime: "14:33",
				Items: []models.Item{
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
				},
				Total: "9.00",
			},
			expected: 109,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := CalculatePoints(testCase.receipt)
			if result != testCase.expected {
				t.Errorf("Result was %v; want %v", result, testCase.expected)
			}
		})
	}
}
