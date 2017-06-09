package verdict

type Repository interface {
	Publish(result []*IpSentence) error
	FetchRules() (result []IpSentence, err error)
	Migrate() error
	Store(*IpSentence) (int64, error)
}
