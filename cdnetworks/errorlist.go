package cdnetworks

const (
	ERR_RATE_LIMIT = "WPLUS_AccountApiTooFrequence"
)

func isAbleToRetry(errCode string) bool {
	switch errCode {
	case ERR_RATE_LIMIT:
		return true
	default:
		return false
	}
}
