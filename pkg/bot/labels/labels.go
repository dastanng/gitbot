package labels

import (
	"encoding/json"
	"io/ioutil"

	"github.com/google/go-github/github"
)

// github labels
const (
	Hold           = "do-not-merge/hold"
	WorkInProgress = "do-not-merge/work-in-progress"
	Approved       = "approved"
	LGTM           = "lgtm"
)

// LoadPresets load preset labels from file
func LoadPresets() ([]*github.Label, error) {

	b, err := ioutil.ReadFile("preset_labels.json")
	if err != nil {
		return nil, err
	}

	var labels []*github.Label

	if err := json.Unmarshal(b, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}
