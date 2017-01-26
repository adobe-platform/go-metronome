`go-metronome` is [semantically versioned](http://semver.org/spec/v2.0.0.html)


### v0.7
- Included undocumented `embed=` uri parameters found in the metronome source that give better picture of status
- Linted code
- Use behance fork of Logrus for compatability with other projects
- Produce docker image that is 19M!


### v0.6
- Implemented metronome v1 API
- Produced cli 


### v0.5
-  Enable job create & run in one step job create ... --run-now'
-  Correctly emit CommandExecute result to stdout based on type.  In particular json.RawMessage which need special handling

### v0.4
- update some doc
 
### v0.3
- add docker compose that stands up test environment
 
### v0.2
- separated source for cli into appropriate files

### v0.1
- Fixed restart cli.
- Replaced Printf with logrus.Debugf
- Added Artifact structure

