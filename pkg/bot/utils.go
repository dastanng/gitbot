package bot

import (
	"context"

	"github.com/golang/glog"

	"github.com/dastanng/gitbot/pkg/bot/labels"
)

func (b *Bot) addPresetLabels(owner, repo string) error {

	recognizedLabels, err := b.getRepoLabels(owner, repo)
	if err != nil {
		glog.Errorf("getRepoLabels err: %v", err)
		return err
	}

	labels, err := labels.LoadPresets()
	if err != nil {
		glog.Errorf("LoadPresets err: %v", err)
		return err
	}

	for _, l := range labels {
		// create preset label if label does not exist
		if _, ok := recognizedLabels[*l.Name]; !ok {
			_, _, err := b.git.Issues.CreateLabel(context.Background(), owner, repo, l)
			if err != nil {
				glog.Errorf("git.Issues.CreateLabel err: %v", err)
				return err
			}
		}
	}

	glog.Infof("add preset labels to %s/%s succeed!", owner, repo)
	return nil
}
