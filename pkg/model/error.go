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

package model

// TerminalExceptionCodes returns terminal exception codes as
// []string for custom resource, if specified in generator config
func (r *CRD) TerminalExceptionCodes() []string {
	if r.cfg == nil {
		return nil
	}
	resGenConfig, found := r.cfg.Resources[r.Names.Original]
	if found && resGenConfig.Exceptions != nil {
		return resGenConfig.Exceptions.TerminalCodes
	}
	return nil
}

// ExceptionCode returns the name of the resource's Exception code for the
// Exception having the exception code. If the generator config has
// instructions for overriding the name of an exception code for a resource for
// a particular HTTP status code, we return that, otherwise we look through the
// API model definitions looking for a match
func (r *CRD) ExceptionCode(httpStatusCode int) string {
	if r.cfg != nil {
		resGenConfig, found := r.cfg.Resources[r.Names.Original]
		if found && resGenConfig.Exceptions != nil {
			if excConfig, present := resGenConfig.Exceptions.Errors[httpStatusCode]; present {
				return excConfig.Code
			}
		}
	}
	if r.Ops.ReadOne != nil {
		op := r.Ops.ReadOne
		for _, errShapeRef := range op.ErrorRefs {
			if errShapeRef.Shape.ErrorInfo.HTTPStatusCode == httpStatusCode {
				code := errShapeRef.Shape.ErrorInfo.Code
				if code != "" {
					return code
				}
				return errShapeRef.Shape.ShapeName
			}
		}
	}
	if r.Ops.ReadMany != nil {
		op := r.Ops.ReadMany
		for _, errShapeRef := range op.ErrorRefs {
			if errShapeRef.Shape.ErrorInfo.HTTPStatusCode == httpStatusCode {
				code := errShapeRef.Shape.ErrorInfo.Code
				if code != "" {
					return code
				}
				return errShapeRef.Shape.ShapeName
			}
		}
	}
	if r.Ops.GetAttributes != nil {
		op := r.Ops.GetAttributes
		for _, errShapeRef := range op.ErrorRefs {
			if errShapeRef.Shape.ErrorInfo.HTTPStatusCode == httpStatusCode {
				code := errShapeRef.Shape.ErrorInfo.Code
				if code != "" {
					return code
				}
				return errShapeRef.Shape.ShapeName
			}
		}
	}
	return "UNKNOWN"
}
