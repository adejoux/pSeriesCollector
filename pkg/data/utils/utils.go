package utils

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
	"time"
)

// WaitAlignForNextCycle waiths untile a next cycle begins aligned with second 00 of each minute
func WaitAlignForNextCycle(SecPeriod int, l *logrus.Logger) {
	i := int64(time.Duration(SecPeriod) * time.Second)
	remain := i - (time.Now().UnixNano() % i)
	l.Infof("Waiting %s to round until nearest interval... (Cycle = %d seconds)", time.Duration(remain).String(), SecPeriod)
	time.Sleep(time.Duration(remain))
}

// RemoveDuplicatesUnordered removes duplicated elements in the array string
func RemoveDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

// DiffSlice return de Difference between two Slices
func DiffSlice(X, Y []string) []string {

	diff := []string{}
	vals := map[string]struct{}{}

	for _, x := range X {
		vals[x] = struct{}{}
	}

	for _, x := range Y {
		if _, ok := vals[x]; !ok {
			diff = append(diff, x)
		}
	}

	return diff
}

func diffKeysInMap(X, Y map[string]string) map[string]string {

	diff := map[string]string{}

	for k, vK := range X {
		if _, ok := Y[k]; !ok {
			diff[k] = vK
		}
	}

	return diff
}

// DiffKeyValuesInMap does a diff key and values from 2 strings maps
func DiffKeyValuesInMap(X, Y map[string]string) map[string]string {

	diff := map[string]string{}

	for kX, vX := range X {
		if vY, ok := Y[kX]; !ok {
			//not exist
			diff[kX] = vX
		} else {
			//exist
			if vX != vY {
				//but value is different
				diff[kX] = vX
			}
		}
	}
	return diff
}

// MapDupAndAdd duplicate map and also add new values
func MapDupAndAdd(source, add map[string]string) map[string]string {
	RetMap := make(map[string]string)
	for k, v := range source {
		RetMap[k] = v
	}
	for k, v := range add {
		RetMap[k] = v
	}
	return RetMap
}

// KeyValArrayToMap return a map from a key value array
func KeyValArrayToMap(keyvalar []string) (map[string]string, error) {
	ret := make(map[string]string)
	if len(keyvalar) > 0 {
		for _, keyval := range keyvalar {
			s := strings.Split(keyval, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				ret[key] = value
			} else {
				return ret, fmt.Errorf("Error on tag definition KEY=VALUE definition [ %s ]", keyval)
			}
		}
	} else {
		return ret, fmt.Errorf("No key value detected ")
	}
	return ret, nil
}

// MapAdd Add map to devices
func MapAdd(dest map[string]string, orig map[string]string) {
	for k, v := range orig {
		dest[k] = v
	}
}
