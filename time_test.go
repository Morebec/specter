package specter

import "time"

func staticTimeProvider(dt time.Time) TimeProvider {
	return func() time.Time {
		return dt
	}
}
