package collector

type E621CollectorEmptyProxy struct{}
type E621CollectorUnknownProxy struct{}
type E621CollectorZeroPoolID struct{}
type E621CollectorNullPoolID struct{}
type E621CollectorEmptyTags struct{}
type E621CollectorNullScrapPages struct{}
type E621CollectorZeroScrapPages struct{}
type E621CollectorNullLogger struct{}
type E621CollectorNullDB struct{}

func (e E621CollectorEmptyProxy) Error() string {
	return "Proxy URL cannot be empty"
}

func (e E621CollectorUnknownProxy) Error() string {
	return "Proxy URL has unknown scheme or protocol"
}

func (e E621CollectorZeroPoolID) Error() string {
	return "Pool ID cannot be zero"
}

func (e E621CollectorNullPoolID) Error() string {
	return "Pool ID cannot be null"
}

func (e E621CollectorEmptyTags) Error() string {
	return "Search tags cannot be empty"
}

func (e E621CollectorNullScrapPages) Error() string {
	return "Maximum scrap pages cannot be null"
}

func (e E621CollectorZeroScrapPages) Error() string {
	return "Maximum scrap pages cannot be zero"
}

func (e E621CollectorNullLogger) Error() string {
	return "Logger cannot be null"
}

func (e E621CollectorNullDB) Error() string {
	return "Database cannot be null"
}
