// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ExecuteWithTimeout takes a function and invokes it, if the function returns false then it will
// retry it after sleeping interval seconds, until it returns true or timeout
func ExecuteWithTimeout(action func() bool, timeout time.Duration, interval time.Duration) error {
	log.Info().Msgf("start for-loop with timeout %v, interval %v", timeout, interval)
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	for {
		select {
		case <-ctx.Done():
			return errors.New("operation timed out")
		default:
			if action() {
				return nil
			}
			time.Sleep(interval)
		}
	}
}

// RetryWithMaxAttempts keeps invoking the passed function until maximum attempts is reached or the
// function returns no error
func RetryWithMaxAttempts(action func() error, attempts int, interval time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Err(err).Msgf("retry (%v remaining) in %v seconds", attempts-i, interval)
			time.Sleep(interval)
		}
		err = action()
		if err == nil {
			return nil
		}
	}
	return errors.Wrapf(err, "failed after %v retries", attempts)
}
