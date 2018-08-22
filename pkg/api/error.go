package api

// Error is a type that is encoded as JSON and returned in an API response in
// the event of an error. Only the Message field is required. More fields may
// be added to this struct in the future for better error reporting.
type Error struct {
	// Message describes the error in English. Typically it is set to
	// `err.Error()`. This field is required.
	Message string `json:"message"`

	// TODO: add a Param field with the (omitempty option in the json tag)
	// to indicate that the error was caused by an invalid, missing, or
	// incorrect parameter. This is not trivial as the API does not
	// currently do parameter validation itself. For example, the
	// /gateway/connect endpoint relies on the gateway.Connect method to
	// validate the netaddress. However, this prevents the API from knowing
	// whether an error returned by gateway.Connect is because of a
	// connection error or an invalid netaddress parameter. Validating
	// parameters in the API is not sufficient, as a parameter's value may
	// be valid or invalid depending on the current state of a module.
}

// Error implements the error interface for the Error type. It returns only the
// Message field.
func (err Error) Error() string {
	return err.Message
}
