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

	// Category is the predefined category ID this reply belongs to.
	// Use one of the PredefinedCategories IDs, or empty for uncategorized.
	Category string `json:"category,omitempty" toml:"category,omitempty"`
}

// PredefinedCategory describes one category in the quick-reply classification system.
type PredefinedCategory struct {
	// ID is the internal identifier stored in QuickReply.Category.
	ID string `json:"id"`

	// NameZh is the Chinese display name.
	NameZh string `json:"name_zh,omitempty"`

	// NameEn is the English display name.
	NameEn string `json:"name_en,omitempty"`
}

// DefaultCategories returns the five project-phase categories recommended
// for classifying quick replies by development stage.
func DefaultCategories() []PredefinedCategory {
	return []PredefinedCategory{
		{ID: "requirements", NameZh: "需求分析", NameEn: "Requirements"},
		{ID: "design", NameZh: "设计", NameEn: "Design"},
		{ID: "development", NameZh: "开发", NameEn: "Development"},
		{ID: "testing", NameZh: "测试", NameEn: "Testing"},
		{ID: "deployment", NameZh: "发布运维", NameEn: "Deployment & Ops"},
	}
}

// ValidCategoryID returns true if the given ID is one of the predefined categories.
func ValidCategoryID(id string) bool {
	for _, c := range DefaultCategories() {
		if c.ID == id {
			return true
		}
	}
	return id == "" // empty is allowed for uncategorized
}

// store is the on-disk structure serialised as TOML.
type store struct {
	QuickReplies []QuickReply `toml:"quick_replies"`
}
