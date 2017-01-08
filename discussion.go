package Esprit

type Discussion struct {
	Class         string
	StartTime     int64
	StopTime      int64
	Contributions []Contribution
	Items         []Item
}

type Contribution struct {
	ID    int
	Type  string
	Color string
}

type Item struct {
	User           string
	Timestamp      int64
	ContributionID int
}
