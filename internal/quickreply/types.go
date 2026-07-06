// Package quickreply manages a user-configurable list of quick-reply templates
// displayed as one-click buttons in the serve web UI. It is a self-contained
// module intended to be maintained as a long-lived local patch with minimal
// merge-conflict surface: it depends only on the standard library and
// BurntSushi/toml, and never imports from internal/serve or internal/config.
package quickreply

// QuickReply is a single preset message the user can send with one click.
type QuickReply struct {
	// Name is the button label shown in the UI.
	Name string `json:"name" toml:"name"`

	// Body is the message text inserted or sent when the button is clicked.
	Body string `json:"body" toml:"body"`

	// AutoSend controls behaviour on click:
	//   true  – send the message immediately (calls /submit)
	//   false – fill the composer input so the user can edit first.
	AutoSend bool `json:"autoSend" toml:"auto_send"`

	// Icon is an optional emoji or short string prepended to the button label.
	Icon string `json:"icon,omitempty" toml:"icon,omitempty"`
}

// store is the on-disk structure serialised as TOML.
type store struct {
	QuickReplies []QuickReply `toml:"quick_replies"`
}
