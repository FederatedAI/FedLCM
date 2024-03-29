// Copyright 2022-2023 VMware, Inc.
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
	FedLCMOpenFLDirector030ChartArchiveContent = getFedLCMOpenFLDirector030ChartArchiveContent()

	FedLCMOpenFLEnvoy030ChartArchiveContent = getFedLCMOpenFLEnvoy030ChartArchiveContent()
)

func getFedLCMOpenFLDirector030ChartArchiveContent() []byte {
	base64Content := `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`
	content, _ := base64.StdEncoding.DecodeString(base64Content)
	return content
}

func getFedLCMOpenFLEnvoy030ChartArchiveContent() []byte {
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
