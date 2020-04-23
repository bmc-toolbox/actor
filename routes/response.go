package routes

// response represents an action response
type response struct {
	Action  string `json:"action"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// errorResponse represents not an action error, i.e. BadRequest, StatusPreconditionFailed
type errorResponse struct {
	Error string `json:"error"`
}

func newResponse(action string, status bool, message string, err error) response {
	resp := response{
		Action:  action,
		Status:  status,
		Message: message,
		Error:   "",
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func newErrorResponse(err error) errorResponse {
	// if err is nil it is a mistake in the code, do not return it as an error to a user
	return errorResponse{
		Error: err.Error(),
	}
}
