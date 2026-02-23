package exceptions

import "fmt"

// ExcelParserConfigurationException represents errors in parser configuration
type ExcelParserConfigurationException struct {
	Message string
	Cause   error
}

func (e ExcelParserConfigurationException) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ExcelParserConfigurationException: %s - %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("ExcelParserConfigurationException: %s", e.Message)
}

func NewExcelParserConfigurationException(message string, cause error) error {
	return ExcelParserConfigurationException{
		Message: message,
		Cause:   cause,
	}
}

// ExcelParserInputException represents errors in input data
type ExcelParserInputException struct {
	Message string
	Cause   error
}

func (e ExcelParserInputException) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ExcelParserInputException: %s - %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("ExcelParserInputException: %s", e.Message)
}

func NewExcelParserInputException(message string, cause error) error {
	return ExcelParserInputException{
		Message: message,
		Cause:   cause,
	}
}

// LayoutMatchException represents errors when no layout matches the header row
type LayoutMatchException struct {
	Message string
}

func (e LayoutMatchException) Error() string {
	return fmt.Sprintf("LayoutMatchException: %s", e.Message)
}

func NewLayoutMatchException(message string) error {
	return LayoutMatchException{
		Message: message,
	}
}

// ExcelParserException represents generic parser errors
type ExcelParserException struct {
	Message string
	Cause   error
}

func (e ExcelParserException) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ExcelParserException: %s - %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("ExcelParserException: %s", e.Message)
}

func NewExcelParserException(message string, cause error) error {
	return ExcelParserException{
		Message: message,
		Cause:   cause,
	}
}
