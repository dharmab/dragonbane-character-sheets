package model

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseQuantity splits "Rope x3" into ("Rope", 3).
func ParseQuantity(s string) (string, int) {
	if i := strings.LastIndex(s, " x"); i >= 0 {
		if n, err := strconv.Atoi(s[i+2:]); err == nil && n >= 2 {
			return s[:i], n
		}
	}
	return s, 1
}

// ApplyQuantity formats a base name with a quantity suffix ("Rope x3")
func ApplyQuantity(name string, quantity int) string {
	if quantity <= 1 {
		return name
	}
	return fmt.Sprintf("%s x%d", name, quantity)
}
