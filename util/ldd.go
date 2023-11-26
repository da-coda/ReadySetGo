package util

import (
	"fmt"
	"github.com/u-root/u-root/pkg/ldd"
)

func IsStaticallyLinkedBinary(file string) (bool, error) {
	infos, err := ldd.Ldd([]string{file})
	if err != nil {
		return false, fmt.Errorf("unable to ldd file: %w", err)
	}

	return len(infos) == 1, nil
}
