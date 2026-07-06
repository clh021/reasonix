// Package quickreply manages a user-configurable list of quick-reply templates
// displayed as one-click buttons in the serve web UI. It is a self-contained
// module intended to be maintained as a long-lived local patch with minimal
// merge-conflict surface: it depends only on the standard library and
// BurntSushi/toml, and never imports from internal/serve or internal/config.
package quickreply

// QuickReply is a single preset message template saved by the user.
type QuickReply struct {
	// Name is the label shown in the quick-reply picker.
	Name string `json:"name" toml:"name"`

	// Body is the message text inserted into the composer when picked.
	Body string `json:"body" toml:"body"`
}

// store is the on-disk structure serialised as TOML.
type store struct {
	QuickReplies []QuickReply `toml:"quick_replies"`
}
