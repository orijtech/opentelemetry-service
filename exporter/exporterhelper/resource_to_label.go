// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporterhelper

import (
	"go.opentelemetry.io/collector/model/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

// ResourceToTelemetrySettings defines configuration for converting resource attributes to metric labels.
type ResourceToTelemetrySettings struct {
	// Enabled indicates whether to not convert resource attributes to metric labels
	Enabled bool `mapstructure:"enabled"`
}

// defaultResourceToTelemetrySettings returns the default settings for ResourceToTelemetrySettings.
func defaultResourceToTelemetrySettings() ResourceToTelemetrySettings {
	return ResourceToTelemetrySettings{
		Enabled: false,
	}
}

// convertResourceToLabels converts all resource attributes to metric labels
func convertResourceToLabels(md pdata.Metrics) pdata.Metrics {
	cloneMd := md.Clone()
	rms := cloneMd.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		resource := rms.At(i).Resource()

		labelMap := extractLabelsFromResource(&resource)

		ilms := rms.At(i).InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metricSlice := ilm.Metrics()
			for k := 0; k < metricSlice.Len(); k++ {
				metric := metricSlice.At(k)
				addLabelsToMetric(&metric, labelMap)
			}
		}
	}
	return cloneMd
}

// extractAttributesFromResource extracts the attributes from a given resource and
// returns them as a StringMap.
func extractLabelsFromResource(resource *pdata.Resource) pdata.StringMap {
	labelMap := pdata.NewStringMap()

	attrMap := resource.Attributes()
	attrMap.Range(func(k string, av pdata.AttributeValue) bool {
		stringLabel := tracetranslator.AttributeValueToString(av)
		labelMap.Upsert(k, stringLabel)
		return true
	})
	return labelMap
}

// addLabelsToMetric adds additional labels to the given metric
func addLabelsToMetric(metric *pdata.Metric, labelMap pdata.StringMap) {
	switch metric.DataType() {
	case pdata.MetricDataTypeGauge:
		addLabelsToNumberDataPoints(metric.Gauge().DataPoints(), labelMap)
	case pdata.MetricDataTypeSum:
		addLabelsToNumberDataPoints(metric.Sum().DataPoints(), labelMap)
	case pdata.MetricDataTypeHistogram:
		addLabelsToDoubleHistogramDataPoints(metric.Histogram().DataPoints(), labelMap)
	}
}

func addLabelsToNumberDataPoints(ps pdata.NumberDataPointSlice, newLabelMap pdata.StringMap) {
	for i := 0; i < ps.Len(); i++ {
		joinStringMaps(newLabelMap, ps.At(i).LabelsMap())
	}
}

func addLabelsToDoubleHistogramDataPoints(ps pdata.HistogramDataPointSlice, newLabelMap pdata.StringMap) {
	for i := 0; i < ps.Len(); i++ {
		joinStringMaps(newLabelMap, ps.At(i).LabelsMap())
	}
}

func joinStringMaps(from, to pdata.StringMap) {
	from.Range(func(k, v string) bool {
		to.Upsert(k, v)
		return true
	})
}
