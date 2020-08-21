package model

import (
	"regexp"
	"strconv"
	"strings"
)

type ValidationHelper struct{}

type IErrInvalidParam interface {
	Message() string
	Error() string
}

/**
This method performs a check if errInvalidParams is strictly caused by ignorableFieldNames.
This means all the ignorableFieldNames should be present in errInvalidParams.

The method returns true if all the validation errors are due to ignorableFieldNames, otherwise false.
*/
func (vh *ValidationHelper) IsValidationErrorIgnorable(ignorableFieldNames []string, errInvalidParams IErrInvalidParam) bool {

	if len(ignorableFieldNames) == 0 {
		return false
	}
	//Invalid parameter exception has predefined message telling total number of errors.
	re, error := regexp.Compile(`(\d+) validation error\(s\) found`)
	if error != nil {
		return false
	} else {
		match := re.FindStringSubmatch(errInvalidParams.Message())
		if len(match) != 2 {
			return false
		}
		validationErrorCount, error := strconv.Atoi(match[1])
		if error != nil {
			return false
		} else if validationErrorCount != len(ignorableFieldNames) { //strict check
			return false
		} else {
			//Make sure all the requiredStatusFieldNames are present in the validationError
			allIgnorableFieldsPresent := true
			for _, ignorableFieldName := range ignorableFieldNames {
				allIgnorableFieldsPresent = allIgnorableFieldsPresent && strings.Contains(strings.ToLower(errInvalidParams.Error()), strings.ToLower(ignorableFieldName))
			}
			if !allIgnorableFieldsPresent {
				return false
			}
		}
	}
	return true
}
