/*
Copyright 2018 Google LLC. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"errors"
	"reflect"

	"github.com/golang/glog"
	"github.com/google/elafros/pkg/apis/ela/v1alpha1"
	"github.com/mattbaird/jsonpatch"
)

var (
	errInvalidRevisions        = errors.New("The route must has exactly one of revision or configuration in traffic field")
	errInvalidRouteInput       = errors.New("Failed to convert input into Route")
	errInvalidTargetPercentSum = errors.New("The route must has traffic percent sum equal to 100")
	errNegativeTargetPercent   = errors.New("The route cannot has negative traffic percent")
	errNonEmptyStatusInRoute   = errors.New("The route cannot have status when it is created")
)

// ValidateRoute is Route resource specific validation and mutation handler
func ValidateRoute(patches *[]jsonpatch.JsonPatchOperation, old GenericCRD, new GenericCRD) error {
	var oldRoute *v1alpha1.Route
	if old != nil {
		var ok bool
		oldRoute, ok = old.(*v1alpha1.Route)
		if !ok {
			return errInvalidRouteInput
		}
	}
	glog.Infof("ValidateRoute: OLD Route is\n%+v", oldRoute)
	newRoute, ok := new.(*v1alpha1.Route)
	if !ok {
		return errInvalidRouteInput
	}
	glog.Infof("ValidateRoute: NEW Route is\n%+v", newRoute)

	if err := validateRoute(newRoute); err != nil {
		return err
	}

	return nil
}

func validateRoute(route *v1alpha1.Route) error {
	if !reflect.DeepEqual(route.Status, v1alpha1.RouteStatus{}) {
		return errNonEmptyStatusInRoute
	}

	return validateTrafficTarget(route)
}

func validateTrafficTarget(route *v1alpha1.Route) error {
	// A service as a placeholder that's not backed by anything is allowed.
	if route.Spec.Traffic == nil {
		return nil
	}

	percentSum := 0
	for _, trafficTarget := range route.Spec.Traffic {
		revisionLen := len(trafficTarget.Revision)
		configurationLen := len(trafficTarget.Configuration)
		if (revisionLen == 0 && configurationLen == 0) ||
			(revisionLen != 0 && configurationLen != 0) {
			return errInvalidRevisions
		}

		if trafficTarget.Percent < 0 {
			return errNegativeTargetPercent
		}
		percentSum += trafficTarget.Percent
	}

	if percentSum != 100 {
		return errInvalidTargetPercentSum
	}
	return nil
}