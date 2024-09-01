package specter

import "time"

type TimeProvider func() time.Time

func CurrentTimeProvider() TimeProvider {
	return func() time.Time {
		return time.Now()
	}
}
