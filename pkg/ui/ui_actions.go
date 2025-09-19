package ui

import (
	"time"
)

// UIActionType represents the type of UI action
type UIActionType string

const (
	ActionClick           UIActionType = "click"
	ActionType            UIActionType = "type"
	ActionNavigate        UIActionType = "navigate"
	ActionWait            UIActionType = "wait"
	ActionScrollToElement UIActionType = "scroll_to_element"
	ActionScrollPage      UIActionType = "scroll_page"
	ActionHover           UIActionType = "hover"
	ActionSelect          UIActionType = "select"
	ActionClear           UIActionType = "clear"
	ActionSubmit          UIActionType = "submit"
	ActionRefresh         UIActionType = "refresh"
	ActionGoBack          UIActionType = "go_back"
	ActionGoForward       UIActionType = "go_forward"
	// Mobile-specific actions
	ActionTap            UIActionType = "tap"
	ActionSwipe          UIActionType = "swipe"
	ActionSwipeLeft      UIActionType = "swipe_left"
	ActionSwipeRight     UIActionType = "swipe_right"
	ActionSwipeUp        UIActionType = "swipe_up"
	ActionSwipeDown      UIActionType = "swipe_down"
	ActionLongPress      UIActionType = "long_press"
	ActionPinch          UIActionType = "pinch"
	ActionSetOrientation UIActionType = "set_orientation"
)

// UIActionOptions holds additional options for UI actions
type UIActionOptions struct {
	Timeout       time.Duration          `json:"timeout,omitempty"`
	WaitCondition string                 `json:"wait_condition,omitempty"`
	ScrollOptions *ScrollOptions         `json:"scroll_options,omitempty"`
	SelectOptions *SelectOptions         `json:"select_options,omitempty"`
	ClickOptions  *ClickOptions          `json:"click_options,omitempty"`
	TypeOptions   *TypeOptions           `json:"type_options,omitempty"`
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// ScrollOptions holds options for scroll actions
type ScrollOptions struct {
	Direction string `json:"direction"` // up, down, left, right, top, bottom
	Distance  int    `json:"distance,omitempty"`
	Speed     string `json:"speed,omitempty"` // slow, normal, fast
}

// SelectOptions holds options for select actions
type SelectOptions struct {
	ByValue bool `json:"by_value,omitempty"`
	ByText  bool `json:"by_text,omitempty"`
	ByIndex bool `json:"by_index,omitempty"`
}

// ClickOptions holds options for click actions
type ClickOptions struct {
	DoubleClick bool `json:"double_click,omitempty"`
	RightClick  bool `json:"right_click,omitempty"`
	Force       bool `json:"force,omitempty"`
}

// TypeOptions holds options for type actions
type TypeOptions struct {
	ClearFirst bool          `json:"clear_first,omitempty"`
	Delay      time.Duration `json:"delay,omitempty"`
}
