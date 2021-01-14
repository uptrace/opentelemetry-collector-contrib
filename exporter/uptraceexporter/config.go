// Copyright 2019 OpenTelemetry Authors
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

package uptraceexporter

import "go.opentelemetry.io/collector/config/configmodels"

type Config struct {
	configmodels.ExporterSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.

	// DSN is a data source name for your Uptrace project.
	// Example: https://<key>@api.uptrace.dev/<project_id>
	DSN string `mapstructure:"dsn"`

	// MaxBatchSize is the maximum number of spans to send in a single batch.
	// The default is 5000.
	MaxBatchSize int `mapstructure:"max_batch_size"`
}
