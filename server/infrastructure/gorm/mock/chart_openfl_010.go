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

package mock

import "encoding/base64"

var (
	FedLCMOpenFLDirector010ChartArchiveContent = getFedLCMOpenFLDirector010ChartArchiveContent()

	FedLCMOpenFLEnvoy010ChartArchiveContent = getFedLCMOpenFLEnvoy010ChartArchiveContent()
)

func getFedLCMOpenFLDirector010ChartArchiveContent() []byte {
	base64Content := `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`
	content, _ := base64.StdEncoding.DecodeString(base64Content)
	return content
}

func getFedLCMOpenFLEnvoy010ChartArchiveContent() []byte {
	base64Content := `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`
	content, _ := base64.StdEncoding.DecodeString(base64Content)
	return content
}

const (
	DefaultEnvoyConfig = `
shard_descriptor:
  template: dummy_shard_descriptor.FedLCMDummyShardDescriptor`
)

var DefaultEnvoyPythonConfig = map[string]string{"dummy_shard_descriptor.py": `
from typing import Iterable
from typing import List
import logging

from openfl.interface.interactive_api.shard_descriptor import DummyShardDescriptor

logger = logging.getLogger(__name__)

class FedLCMDummyShardDescriptor(DummyShardDescriptor):
    """Dummy shard descriptor class."""

    def __init__(self) -> None:
        """Initialize DummyShardDescriptor."""
        super().__init__(['1'], ['1'], 128)

    @property
    def sample_shape(self) -> List[str]:
        """Return the sample shape info."""
        return ['1']

    @property
    def target_shape(self) -> List[str]:
        """Return the target shape info."""
        return ['1']

    @property
    def dataset_description(self) -> str:
        logger.info(
            f'Sample shape: {self.sample_shape}, '
            f'target shape: {self.target_shape}'
        )
        """Return the dataset description."""
        return 'This is dummy data shard descriptor provided by FedLCM project. You should implement your own data ' \
               'loader to load your local data for each Envoy.'
`}
