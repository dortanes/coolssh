package session

// PromptKind is what a session asks the user.
type PromptKind string

const (
	// HostKeyPrompt asks whether to trust a host key neither known_hosts file holds.
	HostKeyPrompt PromptKind = "host-key"
	// PasswordPrompt asks for a host's password, which the user may ask to save.
	PasswordPrompt PromptKind = "password"
	// PassphrasePrompt asks for an encrypted key's passphrase.
	PassphrasePrompt PromptKind = "passphrase"
	// QuestionsPrompt asks the server's keyboard-interactive questions.
	QuestionsPrompt PromptKind = "questions"
)

// Prompt is what a session waits for from the user.
type Prompt struct {
	ID   int        `json:"id"`
	Kind PromptKind `json:"kind"`
	// Host is the name of the host the prompt is about: the session's host or a jump host.
	Host string `json:"host"`
	// Retry is true when the previous answer to the same question did not work.
	Retry bool `json:"retry"`
	// Savable is true when a PasswordPrompt's password may be saved, on a host in the library.
	Savable bool `json:"savable,omitempty"`
	// KeyType and Fingerprint describe the host key a HostKeyPrompt is about.
	KeyType     string `json:"keyType,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	// Key names the key a PassphrasePrompt opens.
	Key string `json:"key,omitempty"`
	// Name, Instruction and Questions are a QuestionsPrompt's, as the server sent them.
	Name        string     `json:"name,omitempty"`
	Instruction string     `json:"instruction,omitempty"`
	Questions   []Question `json:"questions,omitempty"`
}

// Question is one of the server's keyboard-interactive questions.
type Question struct {
	Text string `json:"text"`
	// Echo is true when the answer may be shown as it is typed.
	Echo bool `json:"echo"`
}

// Answer is the user's reply to a Prompt.
type Answer struct {
	// Accept is false when the user declined; the session then closes.
	Accept bool `json:"accept"`
	// Values holds the password or passphrase, or an answer per question.
	Values []string `json:"values"`
	// Remember asks to save a typed password on the host once the server accepts it.
	Remember bool `json:"remember"`
}
