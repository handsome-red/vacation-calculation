package get_quota

type Query struct {
	UserID string
}

type Quota struct {
	Total   int
	Blocked int
}

type YearQuota struct {
	Year      string // Рабочий год
	Base      Quota
	Seniority Quota
	Irregular Quota
}

type Result struct {
	Years []YearQuota
}
