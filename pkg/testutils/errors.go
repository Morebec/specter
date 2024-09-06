package testutils

import (
	"github.com/morebec/go-errors/errors"
	"github.com/stretchr/testify/require"
)

func RequireErrorWithCode(c string) require.ErrorAssertionFunc {
	return func(t require.TestingT, err error, i ...interface{}) {
		require.Error(t, err)

		var sysError errors.SystemError
		if !errors.As(err, &sysError) {
			t.Errorf("expected a system error with code %q but got %s", c, err)
		}
		require.Equal(t, c, sysError.Code())
	}
}
