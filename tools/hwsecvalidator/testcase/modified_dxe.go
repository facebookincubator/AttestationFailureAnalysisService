// Copyright 2023 Meta Platforms, Inc. and affiliates.
//
// Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package testcase

import (
	"context"
	"fmt"

	"github.com/immune-gmbh/AttestationFailureAnalysisService/tools/hwsecvalidator/testcase/errors"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/tools/hwsecvalidator/testcase/types"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/tools/hwsecvalidator/testcase/uefiedit"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/tools/hwsecvalidator/testcase/validator"

	ffsConsts "github.com/9elements/converged-security-suite/v2/pkg/uefi/ffs/consts"
	"github.com/klauspost/cpuid/v2"
)

// ModifiedDXE a test case where we imitate compromised DXE by introducing
// benign changes (like changing order of files and/or recompressing)
type ModifiedDXE struct{}

// Setup implements TestCase.
func (t ModifiedDXE) Setup(ctx context.Context, image []byte) error {
	err := uefiedit.InjectBenignVolumeChange(image, 0, ffsConsts.GUIDDXEContainer, ffsConsts.GUIDDXE)
	if err != nil {
		return fmt.Errorf("unable to inject a benign corruption into DXE: %w", err)
	}
	return nil
}

// Matches implements TestCase
func (ModifiedDXE) Matches(fwInfo types.FirmwareInfoProvider) bool {
	return false
	isIntel, err := types.IsArchitecture(fwInfo, cpuid.Intel)
	if err != nil {
		panic(fmt.Sprintf("cannot determine if the architecture is Intel: %v", err))
	}
	return isIntel
}

// Validate implements TestCase.
func (t ModifiedDXE) Validate(ctx context.Context, origImage []byte, opts ...types.Option) error {
	info, err := validator.GetValidationInfo(ctx, t, origImage, opts)
	if err != nil {
		return errors.ErrValidationInfo{Err: err}
	}

	return validator.CommonHostBootUpExpected().Validate(ctx, info)
}

// Severity implements TestCase.
func (ModifiedDXE) Severity() types.Severity {
	return types.SeverityBlocker
}

var _ types.TestCase = ModifiedDXE{}
