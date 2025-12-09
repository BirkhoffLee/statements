package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// Category constants
const (
	CategoryAll        = "All"
	CategoryTransport  = "Transport"
	CategoryFood       = "Food"
	CategoryShopping   = "Shopping"
	CategoryTravel     = "Travel"
	CategoryUtilities  = "Utilities"
	CategoryApplePay   = "ApplePay"
	CategoryPayPal     = "PayPal"
	CategoryForeignFee = "ForeignFee"
	CategoryOther      = "Other"
)

// ToCDB converts full-width characters to half-width characters
func ToCDB(str string) string {
	var result strings.Builder
	for _, r := range str {
		code := int(r)
		if code == 12288 {
			result.WriteRune(rune(code - 12256))
		} else if code > 65280 && code < 65375 {
			result.WriteRune(rune(code - 65248))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GetCleanDescription removes payment provider prefixes from description
func GetCleanDescription(normalizedDesc string) string {
	// Remove APE or APExxxx (4 digits) prefix
	applePayRegex := regexp.MustCompile(`^APE\d{0,4}`)
	cleanDesc := applePayRegex.ReplaceAllString(normalizedDesc, "")

	// Remove LINE Pay prefix (連加*)
	if strings.HasPrefix(cleanDesc, "連加*") {
		cleanDesc = strings.TrimPrefix(cleanDesc, "連加*")
	}

	// Remove Jkopay prefix (街口電支-)
	if strings.HasPrefix(cleanDesc, "街口電支-") {
		cleanDesc = strings.TrimPrefix(cleanDesc, "街口電支-")
	}

	return cleanDesc
}

// DetectDetailedCategory detects granular category based on transaction description
func DetectDetailedCategory(normalizedDesc string) string {
	desc := strings.ToUpper(normalizedDesc)

	// Transportation
	transportPrefixes := []string{"LIME", "UBER", "UBR*", "FREENOW", "ZITY", "FLIXBUS", "FNM*", "FNM ", "TRENITALIA", "TRENORD", "TRAIN", "SCOOTER", "RAILWAY"}
	for _, prefix := range transportPrefixes {
		if strings.HasPrefix(desc, prefix) {
			return CategoryTransport
		}
	}

	// Food & Groceries
	foodPrefixes := []string{
		"CIBO", "PANIFICIO", "DELIVEROO", "CAFE", "CAFFE", "MACELLERIA",
		"ESSELUNGA", "MERCATO", "MERCADO", "RISTORANTE", "RESTAURANT", "OSTERIA", "GELATERIA", "GELATO", "GELATI", "PIZZA", "PIZZERIA",
		"BURGER", "CONAD", "CARREFOUR", "EATALY", "BAR", "TRATTORIA", "DM-",
		"GLOVO", "KFC", "MCDONALDS", "NESPRESSO", "PASTICCERIA",
		"PRETAMANGER", "FIVEGUYS", "AUTOGRILL", "STARBUCKS", "DRINK",
	}
	for _, prefix := range foodPrefixes {
		if strings.Contains(desc, prefix) {
			return CategoryFood
		}
	}

	// Shopping
	shoppingPrefixes := []string{"AMAZON*", "WWW.AMAZON", "DECATHLON", "BRICOCENTER", "TIGROS", "TEMU.COM", "UNIQLO"}
	for _, prefix := range shoppingPrefixes {
		if strings.Contains(desc, prefix) {
			return CategoryShopping
		}
	}

	// Travel & Accommodation
	travelPrefixes := []string{"AIRBNB", "ALBERGO", "AIRPORT", "EASYJET", "TRIP.COM", "EVAAIR", "RYANAIR", "FLYSCOOT", "GOTOGATE", "BOOKINGCOM", "HOTEL", "KKDAY", "KIWICOM"}
	for _, prefix := range travelPrefixes {
		if strings.Contains(desc, prefix) {
			return CategoryTravel
		}
	}

	// Utilities & Services
	utilitiesPrefixes := []string{"APPLE.COM", "VODAFONE", "AWS", "AMAZONWEBSERVICES", "1PASSWORD", "POLITECNICO", "POSTEITALIA", "OPENAI", "POLISPORTIVA", "PORKBUN"}
	for _, prefix := range utilitiesPrefixes {
		if strings.Contains(desc, prefix) {
			return CategoryUtilities
		}
	}

	return CategoryOther
}

// LoadStatements loads statements from a JSON file
func LoadStatements(filename string) ([]Statement, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var statements []Statement
	err = json.Unmarshal(data, &statements)
	if err != nil {
		return nil, err
	}

	return statements, nil
}

// CategorizeTransactions categorizes all transactions from statements
func CategorizeTransactions(statements []Statement) CategorizedTransactions {
	categorized := CategorizedTransactions{
		ApplePay:    make([]Transaction, 0),
		PayPal:      make([]Transaction, 0),
		LinePay:     make([]Transaction, 0),
		Jkopay:      make([]Transaction, 0),
		ForeignFees: make([]Transaction, 0),
		Other:       make([]Transaction, 0),
	}

	applePayRegex := regexp.MustCompile(`^APE(\d{4})`)

	for i := range statements {
		for j := range statements[i].Transactions {
			tx := &statements[i].Transactions[j]

			// Convert description to half-width
			normalizedDesc := ToCDB(tx.Description)
			tx.NormalizedDescription = normalizedDesc

			// Check for Apple Pay and extract metadata
			if strings.HasPrefix(normalizedDesc, "APE") {
				// Extract card last 4 digits if present
				matches := applePayRegex.FindStringSubmatch(normalizedDesc)
				if len(matches) > 1 {
					tx.ApplePayCardLast4 = matches[1]
				}
			}

			// Detect detailed category based on clean description
			cleanDesc := GetCleanDescription(normalizedDesc)
			detailedCategory := DetectDetailedCategory(cleanDesc)

			// Set the category field to the detailed category
			tx.Category = detailedCategory

			// Keep old categorization for summary view compatibility
			if strings.HasPrefix(normalizedDesc, "APE") {
				categorized.ApplePay = append(categorized.ApplePay, *tx)
			} else if strings.HasPrefix(normalizedDesc, "PAYPAL*") || strings.HasPrefix(normalizedDesc, "PP*") {
				categorized.PayPal = append(categorized.PayPal, *tx)
			} else if strings.HasPrefix(normalizedDesc, "連加*") {
				categorized.LinePay = append(categorized.LinePay, *tx)
			} else if strings.HasPrefix(normalizedDesc, "街口電支-") {
				categorized.Jkopay = append(categorized.Jkopay, *tx)
			} else if strings.HasPrefix(normalizedDesc, "國外交易手續費") {
				categorized.ForeignFees = append(categorized.ForeignFees, *tx)
			} else {
				categorized.Other = append(categorized.Other, *tx)
			}
		}
	}

	return categorized
}
