package services

import (
	"strings"
)

// NormalizeText cleans text for matching by removing common stopwords and special chars
func NormalizeText(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")

	// Remove common prefixes/suffixes
	stopwords := []string{
		"sarl", "sas", "eurl", "sci", "sa",
		"prlv", "virement", "vir", "prelevement", "paiement", "carte", "cb", "facture", "ref",
		"le", "la", "les", "de", "du", "des", "au", "aux", "et", "ou",
	}

	words := strings.Fields(text)
	var meaningfulWords []string

	for _, w := range words {
		if len(w) < 2 {
			continue
		} // Skip single char
		isStop := false
		for _, stop := range stopwords {
			if w == stop {
				isStop = true
				break
			}
		}
		if !isStop {
			meaningfulWords = append(meaningfulWords, w)
		}
	}

	return strings.Join(meaningfulWords, " ")
}

// CalculateClientMatchScore determines if a client likely matches a label
// Returns true if significant match found
func CalculateClientMatchScore(clientName, bankLabel string) bool {
	// 1. Direct Containment (Normalized)
	normClient := NormalizeText(clientName)
	normLabel := NormalizeText(bankLabel)

	if normClient == "" || normLabel == "" {
		return false
	}

	// Exact normalized match
	if strings.Contains(normLabel, normClient) {
		return true
	}

	// 2. Token Overlap
	// If "Restaurant Au Coin" (restaurant, coin) vs "Carte Restaurant" (restaurant)
	// We want to match if significant tokens match.
	clientTokens := strings.Fields(normClient)
	labelTokens := strings.Fields(normLabel)

	for _, cToken := range clientTokens {
		// Only consider significant tokens (>2 chars)
		if len(cToken) <= 2 {
			continue
		}

		for _, lToken := range labelTokens {
			if cToken == lToken {
				return true // Found a significant matching word
			}
		}
	}

	return false
}
