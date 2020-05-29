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

package requeue_test

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-service-operator-k8s/pkg/requeue"
)

func TestRequeueNeeded(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name       string
		args       args
		wantErr    string
		wantUnwrap error
	}{
		{
			name: "wraps non-nil error",
			args: args{
				err: errors.New("some error"),
			},
			wantErr:    "some error",
			wantUnwrap: errors.New("some error"),
		},
		{
			name: "wraps nil error",
			args: args{
				err: nil,
			},
			wantErr:    "",
			wantUnwrap: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requeue.Needed(tt.args.err)
			assert.Equal(t, tt.wantErr, got.Error())
			if tt.wantUnwrap != nil {
				assert.EqualError(t, got.Unwrap(), tt.wantUnwrap.Error())
			} else {
				assert.NoError(t, got.Unwrap())
			}
		})
	}
}

func TestRequeueNeededAfter(t *testing.T) {
	type args struct {
		err      error
		duration time.Duration
	}
	tests := []struct {
		name         string
		args         args
		wantErr      string
		wantUnwrap   error
		wantDuration time.Duration
	}{
		{
			name: "wraps non-nil error",
			args: args{
				err:      errors.New("some error"),
				duration: 3 * time.Second,
			},
			wantErr:      "some error",
			wantUnwrap:   errors.New("some error"),
			wantDuration: 3 * time.Second,
		},
		{
			name: "wraps nil error",
			args: args{
				err:      nil,
				duration: 3 * time.Second,
			},
			wantErr:      "",
			wantUnwrap:   nil,
			wantDuration: 3 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requeue.NeededAfter(tt.args.err, tt.args.duration)
			assert.Equal(t, tt.wantErr, got.Error())
			if tt.wantUnwrap != nil {
				assert.EqualError(t, got.Unwrap(), tt.wantUnwrap.Error())
			} else {
				assert.NoError(t, got.Unwrap())
			}
			assert.Equal(t, 3*time.Second, got.Duration())
		})
	}
}
