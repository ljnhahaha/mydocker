package utils

import (
	"fmt"
	"os"
)

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("can not judge if %s exists, err: %v", path, err)
}
