package ui

type NodeInputAttributes struct {
	Type  string
	Name  string
	Label string
	Value string
}

type WizardText struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Id      int64                  `json:"id"`
	Text    string                 `json:"text"`
	Type    string                 `json:"type"`
}

type WizardNodeAnchorAttributes struct {
	Href     string     `json:"href"`
	Id       string     `json:"id"`
	NodeType string     `json:"node_type"`
	Title    WizardText `json:"title"`
}

type WizardNodeImageAttributes struct {
	Height   int64  `json:"height"`
	Id       string `json:"id"`
	NodeType string `json:"node_type"`
	Src      string `json:"src"`
	Width    int64  `json:"width"`
}

type WizardNodeInputAttributes struct {
	Autocomplete *string     `json:"autocomplete,omitempty"`
	Disabled     bool        `json:"disabled"`
	Label        *WizardText `json:"label,omitempty"`
	Name         string      `json:"name"`
	NodeType     string      `json:"node_type"`
	Onclick      *string     `json:"onclick,omitempty"`
	Pattern      *string     `json:"pattern,omitempty"`
	Required     *bool       `json:"required,omitempty"`
	Type         string      `json:"type"`
	Value        interface{} `json:"value,omitempty"`
}

type WizardNodeScriptAttributes struct {
	Async          bool   `json:"async"`
	Crossorigin    string `json:"crossorigin"`
	Id             string `json:"id"`
	Integrity      string `json:"integrity"`
	NodeType       string `json:"node_type"`
	Nonce          string `json:"nonce"`
	Referrerpolicy string `json:"referrerpolicy"`
	Src            string `json:"src"`
	Type           string `json:"type"`
}

type WizardNodeTextAttributes struct {
	Id       string     `json:"id"`
	NodeType string     `json:"node_type"`
	Text     WizardText `json:"text"`
}

type WizardNodeAttributes struct {
	UiNodeAnchorAttributes *WizardNodeAnchorAttributes
	UiNodeImageAttributes  *WizardNodeImageAttributes
	UiNodeInputAttributes  *WizardNodeInputAttributes
	UiNodeScriptAttributes *WizardNodeScriptAttributes
	UiNodeTextAttributes   *WizardNodeTextAttributes
}

type WizardNodeMeta struct {
	Label *WizardText `json:"label,omitempty"`
}

type WizardNode struct {
	Attributes WizardNodeAttributes
	Group      string
	Meta       WizardNodeMeta
	Type       string
}

type WizardPage struct {
	Method   string
	Action   string
	Messages []WizardText
	Nodes    []WizardNode
}

type WizardPages map[string]WizardPage
