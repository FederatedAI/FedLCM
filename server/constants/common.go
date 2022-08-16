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

package constants

import "github.com/pkg/errors"

const APIVersion = "v1"

var (
	// Branch is the source branch
	Branch string

	// Commit is the commit number
	Commit string

	// BuildTime is the compiling time
	BuildTime string
)

var ErrNotImplemented = errors.New("not implemented")
