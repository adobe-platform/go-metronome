package metronome_test

import (
	"testing"

	. "github.com/adobe-platform/go-metronome/metronome"
	//"time"
	//"github.com/ChannelMeter/iso8601duration"

	//"github.com/ChannelMeter/iso8601duration"
	//"github.com/stretchr/testify/assert"
	//. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"


)

func TestImmediateSched(t *testing.T) {
	t.Parallel()
	sched,err := ImmediateSchedule()
	Expect(err,nil)
	Expect(sched.Cron,"* * * * *")

}
