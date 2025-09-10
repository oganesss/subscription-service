package models

type Errors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type ErrorResponse struct {
	Errors Errors `json:"errors"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
