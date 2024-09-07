// Copyright 2024 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
