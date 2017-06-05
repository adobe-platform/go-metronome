package metronome

// A Config defines a client configuration
type Config struct {
	/* the url for metronome */
	URL string
	/* switch on debugging */
	Debug bool
	/* the timeout for requests */
	RequestTimeout int
	/* allow unverified tls (self-signed certs) defaults to false */
	AllowUnverifiedTLS bool

	AuthToken string
	User      string
	Pw        string
}

// NewDefaultConfig returns a default configuration.
// Helpful for local testing/development.
func NewDefaultConfig() Config {
	return Config{
		URL:            "http://127.0.0.1:9000",
		Debug:          false,
		RequestTimeout: 5}
}
