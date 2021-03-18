/*
Copyright © 2021 Bob Callaway <bcallawa@redhat.com>

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
package rpm

import (
	"errors"
	"fmt"

	"github.com/sigstore/rekor/pkg/types"
	"github.com/sigstore/rekor/pkg/util"

	"github.com/go-openapi/swag"
	"github.com/sigstore/rekor/pkg/generated/models"
)

const (
	KIND = "rpm"
)

type BaseRPMType struct{}

func (rt BaseRPMType) Kind() string {
	return KIND
}

func init() {
	types.TypeMap.Set(KIND, New)
}

func New() types.TypeImpl {
	return &BaseRPMType{}
}

var SemVerToFacFnMap = &util.VersionFactoryMap{VersionFactories: make(map[string]util.VersionFactory)}

func (rt BaseRPMType) UnmarshalEntry(pe models.ProposedEntry) (types.EntryImpl, error) {
	rpm, ok := pe.(*models.Rpm)
	if !ok {
		return nil, errors.New("cannot unmarshal non-RPM types")
	}

	if genFn, found := SemVerToFacFnMap.Get(swag.StringValue(rpm.APIVersion)); found {
		entry := genFn()
		if entry == nil {
			return nil, fmt.Errorf("failure generating RPM object for version '%v'", rpm.APIVersion)
		}
		if err := entry.Unmarshal(rpm); err != nil {
			return nil, err
		}
		return entry, nil
	}
	return nil, fmt.Errorf("RPMType implementation for version '%v' not found", swag.StringValue(rpm.APIVersion))
}
