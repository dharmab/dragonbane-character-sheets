package character

import (
	"strconv"
	"strings"
)

// ParseQty splits "Rope x3" into ("Rope", 3). Returns qty=1 when no suffix.
func ParseQty(name string) (base string, qty int) {
	if i := strings.LastIndex(name, " x"); i >= 0 {
		if n, err := strconv.Atoi(name[i+2:]); err == nil && n >= 2 {
			return name[:i], n
		}
	}
	return name, 1
}

// ApplyQty formats a base name with a quantity suffix ("Rope x3"), omitting the
// suffix when qty <= 1.
func ApplyQty(base string, qty int) string {
	if qty <= 1 {
		return base
	}
	return base + " x" + strconv.Itoa(qty)
}
