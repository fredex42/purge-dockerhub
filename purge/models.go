package purge

import (
	"fmt"
	"github.com/fredex42/purge-dockerhub/dockerhub"
)

type DecisionInfo struct {
	TagSpec            *dockerhub.TagSpec
	LastPulledOverTime bool
	LastPushedOverTime bool
	OutsideKeepRecent  bool
}

func (d *DecisionInfo) Dump() string {
	result := "keep"
	if d.CanDelete() {
		result = "delete"
	}
	return fmt.Sprintf("%s: last pulled over? %t last pushed over? %t outside keep recent window? %t result? %s",
		d.TagSpec.GetCanonicalName(), d.LastPulledOverTime, d.LastPushedOverTime, d.OutsideKeepRecent, result)
}

func (d *DecisionInfo) GetCanonicalName() string {
	return d.TagSpec.GetCanonicalName()
}

/**
If all the booleans are true (i.e. it was last pulled a long time ago, last pushed a long time ago AND it's outside the horizon)
then it can be deleted.
*/
func (d *DecisionInfo) CanDelete() bool {
	return d.LastPulledOverTime && d.LastPushedOverTime && d.OutsideKeepRecent
}
