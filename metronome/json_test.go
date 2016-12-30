package metronome_test

import (
	//"time"

	. "github.com/adobe-platform/go-metronome/metronome"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"


	//"fmt"

	"encoding/json"
	/*
	"bytes"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"fmt"
	"net/url"
	"github.com/onsi/gomega/types"
	"strconv"
	*/
)

const data5 = `{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[{"uri":"http://foo.test.com/application.zip","extract":true,"executable":true,"cache":false}],"cmd":"nuke --dry --master local","args":["nuke","--dry","--master","local"],"cpus":1.5,"mem":32,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"MON":"test","CONNECT":"direct"},"maxLaunchDelay":3600,"placement":{"constraints":[{"attribute":"rack","operator":"EQ","value":"rack-2"}]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"/mnt/test","hostPath":"/etc/guest","mode":"RW"}]}}`
const hiddenJobApi = `{"id":"foo.bar","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"echo \"testing $(date)\"","env":{},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"alpine:3.4"},"volumes":[{"containerPath":"/var/lib/tmp","hostPath":"/app","mode":"RO"}],"restart":{"policy":"NEVER","activeDeadlineSeconds":0}},"schedules":[{"id":"every2","cron":"*/2 * * * *","timezone":"Etc/GMT","startingDeadlineSeconds":60,"concurrencyPolicy":"ALLOW","enabled":true,"nextRunAt":"2016-12-12T20:16:00.000+0000"}],"activeRuns":[{"id":"20161212192759dliHA","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:27:59.057+0000","completedAt":null,"tasks":[]},{"id":"20161212192200ouvXM","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:22:00.009+0000","completedAt":null,"tasks":[]},{"id":"20161212200759oAg7W","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:07:59.397+0000","completedAt":null,"tasks":[]},{"id":"20161212194759esoLp","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:47:59.227+0000","completedAt":null,"tasks":[]},{"id":"201612121953592YSlf","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:53:59.278+0000","completedAt":null,"tasks":[]},{"id":"20161212192359aTWwb","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:23:59.026+0000","completedAt":null,"tasks":[]},{"id":"20161212193359ppqF0","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:33:59.101+0000","completedAt":null,"tasks":[]},{"id":"20161212195959c71yS","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:59:59.328+0000","completedAt":null,"tasks":[]},{"id":"20161212193959sxqfe","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:39:59.150+0000","completedAt":null,"tasks":[]},{"id":"20161212200359noaiJ","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:03:59.357+0000","completedAt":null,"tasks":[]},{"id":"20161212191559OFafE","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:15:59.967+0000","completedAt":null,"tasks":[]},{"id":"20161212191759KtwAA","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:17:59.977+0000","completedAt":null,"tasks":[]},{"id":"20161212192559reYwe","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:25:59.049+0000","completedAt":null,"tasks":[]},{"id":"201612121949597k1qV","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:49:59.237+0000","completedAt":null,"tasks":[]},{"id":"20161212193759AsLk8","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:37:59.138+0000","completedAt":null,"tasks":[]},{"id":"20161212200159eu7AS","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:01:59.349+0000","completedAt":null,"tasks":[]},{"id":"20161212190959uY5Gv","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:09:59.921+0000","completedAt":null,"tasks":[]},{"id":"201612121857593BUlr","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:57:59.827+0000","completedAt":null,"tasks":[]},{"id":"20161212195159i31hd","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:51:59.257+0000","completedAt":null,"tasks":[]},{"id":"20161212200559UpGSC","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:05:59.377+0000","completedAt":null,"tasks":[]},{"id":"20161212193559jmPF7","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:35:59.118+0000","completedAt":null,"tasks":[]},{"id":"20161212195759NZBoi","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:57:59.306+0000","completedAt":null,"tasks":[]},{"id":"20161212191359AninT","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:13:59.957+0000","completedAt":null,"tasks":[]},{"id":"201612121859592GqyE","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:59:59.847+0000","completedAt":null,"tasks":[]},{"id":"20161212193159Aqg3w","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:31:59.077+0000","completedAt":null,"tasks":[]},{"id":"20161212190559BKt7A","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:05:59.890+0000","completedAt":null,"tasks":[]},{"id":"20161212195559rJPNh","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:55:59.302+0000","completedAt":null,"tasks":[]},{"id":"20161212190359pVtaU","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:03:59.878+0000","completedAt":null,"tasks":[]},{"id":"20161212201359DtsgH","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:13:59.449+0000","completedAt":null,"tasks":[]},{"id":"20161212194359ikn8A","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:43:59.190+0000","completedAt":null,"tasks":[]},{"id":"20161212191959REe2G","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:19:59.997+0000","completedAt":null,"tasks":[]},{"id":"20161212201159yMUic","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:11:59.428+0000","completedAt":null,"tasks":[]},{"id":"201612121853596eV58","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:53:59.798+0000","completedAt":null,"tasks":[]},{"id":"20161212192959GWEPL","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:29:59.067+0000","completedAt":null,"tasks":[]},{"id":"20161212194559LkryV","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:45:59.213+0000","completedAt":null,"tasks":[]},{"id":"20161212190159Bji8F","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:01:59.858+0000","completedAt":null,"tasks":[]},{"id":"20161212191159WnbI6","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:11:59.937+0000","completedAt":null,"tasks":[]},{"id":"20161212200959c7BoQ","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:09:59.407+0000","completedAt":null,"tasks":[]},{"id":"20161212194159Er4ln","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:41:59.167+0000","completedAt":null,"tasks":[]},{"id":"20161212185559b21Bi","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:55:59.807+0000","completedAt":null,"tasks":[]},{"id":"201612121907597DflY","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:07:59.897+0000","completedAt":null,"tasks":[]}],"history":{"successCount":19,"failureCount":1,"lastSuccessAt":"2016-12-12T18:02:00.251+0000","lastFailureAt":"2016-12-12T17:36:00.409+0000","successfulFinishedRuns":[{"id":"20161212180159qfl4J","createdAt":"2016-12-12T18:01:59.335+0000","finishedAt":"2016-12-12T18:02:00.251+0000"},{"id":"201612121759592WzzE","createdAt":"2016-12-12T17:59:59.314+0000","finishedAt":"2016-12-12T18:00:00.271+0000"},{"id":"20161212175759EuhhS","createdAt":"2016-12-12T17:57:59.294+0000","finishedAt":"2016-12-12T17:58:00.173+0000"},{"id":"20161212175559OsNh4","createdAt":"2016-12-12T17:55:59.277+0000","finishedAt":"2016-12-12T17:56:00.191+0000"},{"id":"201612121753593PwVS","createdAt":"2016-12-12T17:53:59.257+0000","finishedAt":"2016-12-12T17:54:00.387+0000"},{"id":"20161212175159quwN5","createdAt":"2016-12-12T17:51:59.245+0000","finishedAt":"2016-12-12T17:52:00.164+0000"},{"id":"20161212174959jbcuu","createdAt":"2016-12-12T17:49:59.230+0000","finishedAt":"2016-12-12T17:50:00.165+0000"},{"id":"20161212174759EuLBH","createdAt":"2016-12-12T17:47:59.207+0000","finishedAt":"2016-12-12T17:48:00.194+0000"},{"id":"20161212174559Yk4Cc","createdAt":"2016-12-12T17:45:59.186+0000","finishedAt":"2016-12-12T17:46:00.078+0000"},{"id":"20161212174359IdEUB","createdAt":"2016-12-12T17:43:59.164+0000","finishedAt":"2016-12-12T17:44:00.098+0000"},{"id":"20161212174159QbXKn","createdAt":"2016-12-12T17:41:59.154+0000","finishedAt":"2016-12-12T17:42:00.037+0000"},{"id":"201612121739593nv1S","createdAt":"2016-12-12T17:39:59.134+0000","finishedAt":"2016-12-12T17:40:00.114+0000"},{"id":"20161212173759jMH95","createdAt":"2016-12-12T17:37:59.113+0000","finishedAt":"2016-12-12T17:37:59.950+0000"},{"id":"201612121733597QS6t","createdAt":"2016-12-12T17:33:59.743+0000","finishedAt":"2016-12-12T17:34:00.561+0000"},{"id":"201612121731598j283","createdAt":"2016-12-12T17:31:59.733+0000","finishedAt":"2016-12-12T17:32:00.648+0000"},{"id":"20161212172959BqBYh","createdAt":"2016-12-12T17:29:59.715+0000","finishedAt":"2016-12-12T17:30:00.546+0000"},{"id":"20161212172759WGAh2","createdAt":"2016-12-12T17:27:59.694+0000","finishedAt":"2016-12-12T17:28:00.703+0000"},{"id":"20161212172559lbS9Z","createdAt":"2016-12-12T17:25:59.689+0000","finishedAt":"2016-12-12T17:26:00.626+0000"},{"id":"20161212172359E8E0G","createdAt":"2016-12-12T17:23:59.665+0000","finishedAt":"2016-12-12T17:24:00.587+0000"}],"failedFinishedRuns":[{"id":"20161212173559asQpr","createdAt":"2016-12-12T17:35:59.483+0000","finishedAt":"2016-12-12T17:36:00.409+0000"}]},"historySummary":{"successCount":19,"failureCount":1,"lastSuccessAt":"2016-12-12T18:02:00.251+0000","lastFailureAt":"2016-12-12T17:36:00.409+0000"}}`

var _ = Describe("JSON rendering", func() {

	var ()
	It("Makes a job with hidden API field", func() {
		var job2 Job
		err := json.Unmarshal([]byte(hiddenJobApi), &job2)
		Expect(err).NotTo(HaveOccurred())
		Expect(job2.GetRun()).NotTo(Equal(nil))
		Expect(job2.Schedules).ShouldNot(Equal(nil))
		Expect(job2.History).ShouldNot(Equal(nil))
		Expect(job2.HistorySummary).ShouldNot(Equal(nil))
		Expect(job2.HistorySummary.SuccessCount).Should(Equal(19))
		Expect(job2.HistorySummary.FailureCount).Should(Equal(1))
		Expect(len(job2.History.SuccessfulFinishedRuns)).To(Equal(19))
		Expect(len(job2.History.FailedFinishedRuns)).To(Equal(1))

	})
	It("Makes builds a job in Code producing stock API structure", func() {
		runnable, err5 := NewRun(1.5, 32, 128);
		Expect(err5).ShouldNot(HaveOccurred())
		Expect(runnable.Cpus).To(Equal(1.5))
		runnable.SetDocker(&Docker{
			Image: "foo/bla:test",
		}).SetEnv(map[string]string{
			"MON": "test",
			"CONNECT": "direct",
		}).SetArgs([]string{
			"nuke",
			"--dry",
			"--master",
			"local",
		}).SetCmd("nuke --dry --master local").SetPlacement(&Placement{
			Constraints: []Constraint{
				Constraint{Attribute: "rack", Operator: EQ, Value: "rack-2"} },

		}).SetArtifacts([] Artifact{
			Artifact{URI: "http://foo.test.com/application.zip", Extract: true, Executable: true, Cache: false},
		}).SetMaxLaunchDelay(
			3600,
		).SetRestart(&Restart{
			ActiveDeadlineSeconds: 120, Policy: "NEVER",

		}).SetVolumes([]Volume{
			Volume{Mode:RW, HostPath:"/etc/guest", ContainerPath: "/mnt/test", },
		}).SetUser("root")

		job, err6 := NewJob("prod.example.app", "Example Application", Labels{"location":"olympus", "owner":"zeus"}, runnable);
		Expect(err6).NotTo(HaveOccurred())

		_, err := json.Marshal(job)

		Expect(err).NotTo(HaveOccurred())
		// unmarshal precanned
		var job2 Job
		err = json.Unmarshal([]byte(data5), &job2)

		Expect(err).NotTo(HaveOccurred())
		Expect(*job).To(Equal(job2))


	})

})
