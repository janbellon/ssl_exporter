package httpserver

import (
	"fmt"
	"strings"
)

func CheckAuthorization(authorization []string, bearer string) error {
	if len(authorization) == 0 {
		return fmt.Errorf("Missing authorization header")
	}
	parts := strings.SplitN(authorization[0], " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("Authorization must be of type 'bearer <token>'")
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return fmt.Errorf("Authorization must be a bearer token")
	}
	if parts[1] != bearer {
		return fmt.Errorf("Bad token")
	}
	return nil
}
