package rules

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"receipt-processor-challenge-jase180/internal/models"
)

// CalculatePoints computes the total points by calling all rules functions
// Each rule is implemented in own function for separation of concerns
// Functions handle the argument in its original JSON data type, conversion and error handling
func CalculatePoints(receipt models.Receipt) int {
	points := 0

	points += PointsForRetailerName(receipt.Retailer)

	if p, err := PointsForRoundTotal(receipt.Total); err == nil {
		points += p
	}

	if p, err := PointsForQuarterMultiple(receipt.Total); err == nil {
		points += p
	}

	points += PointsForEveryTwoItems(receipt.Items)

	for _, item := range receipt.Items {
		if p, err := PointsForItemDescription(item); err == nil {
			points += p
		}
	}

	if p, err := PointsForOddDay(receipt.PurchaseDate); err == nil {
		points += p
	}

	if p, err := PointsForTimeRange(receipt.PurchaseTime); err == nil {
		points += p
	}

	return points
}

// Rule: One point for every alphanumeric character in the retailer name.
// Utilizes "unicode" to check character for clarity, alternative is range based e.g. c >='a'
func PointsForRetailerName(retailer string) int {
	points := 0
	for _, char := range retailer {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			points++
		}
	}
	return points
}

// Rule: 50 points if the total is a round dollar amount with no cents.
// Pattern provided is "^\\d+\\.\\d{2}$"
// Since there can only be 2 decimals, check by multiply by 100 to avoid floating point errors, alternative is an epsilon
func PointsForRoundTotal(total string) (int, error) {
	// Extra defensive programming to make sure dollar is in pattern provided
	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(total) {
		return 0, nil // fail gracefully and just return 0
	}

	// Convert total from string to float, error handling
	totalFloat, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert total to float: %s", total)
	}

	// Multiply by 100 to get an integer, round to avoid floating point error
	totalCents := int(math.Round(totalFloat * 100))

	if totalCents%100 == 0 {
		return 50, nil
	}
	return 0, nil
}

// Rule: 25 points if the total is a multiple of 0.25. MODULUS STYLE
// simiar to CalculateRoundTotalPoints
func PointsForQuarterMultiple(total string) (int, error) {
	// Extra defensive programming to make sure dollar is in pattern provided
	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(total) {
		return 0, nil // fail gracefully and just return 0
	}
	// Convert total from string to float, error handling
	totalFloat, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert total to float: %s", total)
	}

	// Multiply by 100 to get an integer, round to avoid floating point error
	totalCents := int(math.Round(totalFloat * 100))
	if totalCents%25 == 0 {
		return 25, nil
	}
	return 0, nil
}

// Rule: 5 points for every two items on the receipt.
// Base division handles odd numbers
func PointsForEveryTwoItems(items []models.Item) int {
	return len(items) / 2 * 5
}

// Rule: If the trimmed length of the item description is a multiple of 3,
// multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
func PointsForItemDescription(item models.Item) (int, error) {
	// Trim with "strings" function for simplicity and readability
	trimmedLen := len(strings.TrimSpace(item.ShortDescription))
	if trimmedLen%3 != 0 || trimmedLen == 0 {
		return 0, nil
	}

	priceFloat, err := strconv.ParseFloat(item.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert total to float: %s", item.Price)
	}

	points := int(math.Ceil(priceFloat * 0.2)) // round up to nearest and convert back to int
	return points, nil
}

// Rule: 6 points if the day in the purchase date is odd.
// String slicing not used for better readability, scalability and less error-prone
func PointsForOddDay(purchaseDate string) (int, error) {
	// Parse date with "time" methods
	date, err := time.Parse("2006-01-02", purchaseDate)
	if err != nil {
		return 0, fmt.Errorf("cannot convert date string to date: %s", purchaseDate)
	}

	day := date.Day() // Extract day from date
	if day%2 == 1 {
		return 6, nil
	}
	return 0, nil
}

// Rule: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
// Assume this means range including 14:01 and 15:59
func PointsForTimeRange(purchaseTime string) (int, error) {
	// Parse time with "time" methods
	purchaseTimeParsed, err := time.Parse("15:04", purchaseTime)
	if err != nil {
		return 0, fmt.Errorf("cannot convert time string to time: %s", purchaseTime)
	}

	startTime, _ := time.Parse("15:04", "14:00")
	endTime, _ := time.Parse("15:04", "16:00")

	if purchaseTimeParsed.After(startTime) && purchaseTimeParsed.Before(endTime) {
		return 10, nil
	}
	return 0, nil
}
