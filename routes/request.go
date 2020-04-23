package routes

// request describes the action to be carried out by actor
type request struct {
	ActionSequence []string `json:"action-sequence"`
	CallbackURL    string   `json:"callback-url"`
}
