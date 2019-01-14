package routes

// Request describes the action to be carried out by actor
type Request struct {
	ActionSequence []string `json:"action-sequence"`
	CallbackURL    string   `json:"callback-url"`
}
