package shared

type TimedRequest struct {
	Timestamp           int64
	Source              string
	RequestURI          string
	Target              string
	ProxyServerDuration float64
	TotalDuration       float64
	Status              int
}
