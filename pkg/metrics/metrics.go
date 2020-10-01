// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	outboundAPIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ack_outbound_api_requests_total",
			Help: "Total number of outbound AWS API requests made by the controller.",
		},
		[]string{
			"service",
			"op_type",
			"op_id",
		},
	)
	outboundAPIRequestsErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ack_outbound_api_requests_error_total",
			Help: "Total number of outbound AWS API requests made by the controller that resulted in a 4XX or 5XX HTTP status code.",
		},
		[]string{
			"service",
			"op_id",
			"status_code",
		},
	)
)

func init() {
	prometheus.MustRegister(outboundAPIRequestsTotal)
	prometheus.MustRegister(outboundAPIRequestsErrorTotal)
}

// Metrics contains the set of Prometheus metric objects used to store counter
// and histograms for a variety of data points
type Metrics struct {
	// serviceID is the ID of the AWS service the controller is managing
	serviceID string
	// obAPIRequestTotal contains the total number of outbound AWS API requests
	// made by the service controller
	obAPIRequestTotal *prometheus.CounterVec
	// obAPIRequestErrorTotal contains the total number of outbound AWS API
	// requests made by the service controller that resulted in an HTTP 4XX or
	// 5XX status code
	obAPIRequestErrorTotal *prometheus.CounterVec
}

// RecordAPICall increments appropriate metrics tracking the count and duration
// of a supplied outbound AWS API call
func (m *Metrics) RecordAPICall(
	// The type of API operation, e.g. "CREATE" or "READ_ONE"
	opType string,
	// The name of the AWS API call, e.g. "CreateTopic"
	opID string,
	// The HTTP status code returned from the AWS API call
	statusCode int,
	// Any error that was returned from the aws-sdk-go client call
	err error,
) {
	m.obAPIRequestTotal.With(
		prometheus.Labels{
			"service": m.serviceID,
			"op_type": string(opType),
			"op_id":   opID,
		},
	).Inc()
	if statusCode >= 400 && statusCode < 600 {
		m.obAPIRequestErrorTotal.With(
			prometheus.Labels{
				"service":     m.serviceID,
				"op_id":       opID,
				"status_code": strconv.Itoa(statusCode),
			},
		).Inc()
	}
}

// NewMetrics returns a pointer to a Metrics struct that can be used to collect
// and expose various Prometheus metrics
func NewMetrics(serviceID string) *Metrics {
	return &Metrics{
		serviceID:              serviceID,
		obAPIRequestTotal:      outboundAPIRequestsTotal,
		obAPIRequestErrorTotal: outboundAPIRequestsErrorTotal,
	}
}
