package node

import (
	"errors"
	"fmt"
)

func CompareMaps(m1, m2 map[string]*Node) error {
	if len(m1) != len(m2) {
		return fmt.Errorf("len mismatch %v %v", len(m1), len(m2))
	}
	errs := []error{}

	for k1, v1 := range m1 {
		v2, ok := m2[k1]
		if !ok {
			errs = append(errs, fmt.Errorf("key %v not found", k1))
			continue
		}
		if err := v1.EqualTo(v2); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	fmt.Printf("errors: %v\n", len(errs))

	return errors.Join(errs...)
}
