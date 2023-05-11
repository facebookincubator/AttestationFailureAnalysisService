package thrift

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/facebookincubator/go-belt/tool/logger"

	"libfb/go/go303"
	"libfb/go/stats/export"

	"github.com/immune-gmbh/AttestationFailureAnalysisService/if/afas"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/if/typeconv"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/server/controller"
)

const (
	gcInterval = time.Hour
)

var _ afas.FirmwareAnalyzer = &service{}

type service struct {
	*go303.Base
	Controller *controller.Controller
	Stats      export.Exported
	Cache      *serviceCache
}

func newService(
	base *go303.Base,
	ctrl *controller.Controller,
	stats export.Exported,
) *service {
	return &service{
		Base:       base,
		Controller: ctrl,
		Stats:      stats,
		Cache:      newServiceCache(),
	}
}

func (svc *service) Reset() {
	svc.Cache.Reset()
}

type thriftExceptioner interface {
	ThriftException() error
}

func (svc *service) SearchFirmware(
	ctx context.Context,
	request *afas.SearchFirmwareRequest,
) (report *afas.SearchFirmwareResult_, err error) {
	if request == nil {
		return nil, fmt.Errorf("request == nil")
	}
	report, err = svc.Controller.SearchFirmware(
		ctx,
		request.GetOrFilters(),
		request.GetFetchContent(),
	)
	return report, unwrapException(err)
}

func (svc *service) SearchReport(
	ctx context.Context,
	request *afas.SearchReportRequest,
) (report *afas.SearchReportResult_, err error) {
	if request == nil {
		return nil, fmt.Errorf("request == nil")
	}
	report, err = svc.Controller.SearchReport(
		ctx,
		request.GetOrFilters(),
		uint64(request.GetLimit()),
	)
	return report, unwrapException(err)
}

func (svc *service) ReportHostConfiguration(
	ctx context.Context,
	request *afas.ReportHostConfigurationRequest,
) (*afas.ReportHostConfigurationResult_, error) {
	if request == nil {
		return nil, fmt.Errorf("request == nil")
	}

	// Log entire request except TPMEventLog which could be too long
	loggedRequest := *request
	loggedRequest.TPMEventLog = nil
	logger.FromCtx(ctx).Infof("ReportHostConfiguration with parameters: '%s'", loggedRequest)

	hostTPM, err := typeconv.FromThriftTPMType(request.GetTpmDevice())
	if err != nil {
		return nil, fmt.Errorf("failed to convert TPMType: %w", err)
	}

	pcrs, err := svc.Controller.ReportHostConfiguration(
		ctx,
		request.GetHostInfo(),
		request.GetFirmwareVersion(),
		request.GetFirmwareDateString(),
		hostTPM,
		request.GetStatusRegisters(),
		typeconv.FromThriftTPMEventLog(request.GetTPMEventLog()),
		request.GetPCRValue(),
	)
	if err != nil {
		return nil, unwrapException(err)
	}
	return &afas.ReportHostConfigurationResult_{
		PCR0SHA1:   pcrs.PCR0SHA1,
		PCR0SHA256: pcrs.PCR0SHA256,
	}, nil
}

func (svc *service) Analyze(
	ctx context.Context,
	request *afas.AnalyzeRequest,
) (*afas.AnalyzeResult_, error) {
	if request == nil {
		return nil, fmt.Errorf("request == nil")
	}

	artifacts := make([]afas.Artifact, 0, len(request.GetArtifacts()))
	for idx, art := range request.GetArtifacts() {
		if art == nil {
			return nil, fmt.Errorf("artifact at index '%d' is nil", idx)
		}
		if art.CountSetFieldsArtifact() != 1 {
			return nil, fmt.Errorf("artifact should have exactly 1 value set, but got %d at index %d",
				art.CountSetFieldsArtifact(), idx)
		}
		artifacts = append(artifacts, *art)
	}
	analyzers := make([]afas.AnalyzerInput, 0, len(request.GetAnalyzers()))
	for idx, analyzer := range request.GetAnalyzers() {
		if analyzer == nil {
			return nil, fmt.Errorf("analyzer input at index '%d' is nil", idx)
		}
		if analyzer.CountSetFieldsAnalyzerInput() != 1 {
			return nil, fmt.Errorf("analyzer input should have exactly 1 value set, but got %d at index %d",
				analyzer.CountSetFieldsAnalyzerInput(), idx)
		}
		analyzers = append(analyzers, *analyzer)
	}

	result, err := svc.Controller.Analyze(
		ctx,
		request.GetHostInfo(),
		artifacts,
		analyzers,
	)
	if err != nil {
		return nil, unwrapException(err)
	}
	return result, nil
}

func (svc *service) CheckFirmwareVersion(
	ctx context.Context,
	request *afas.CheckFirmwareVersionRequest,
) (*afas.CheckFirmwareVersionResult_, error) {
	if request == nil {
		return nil, fmt.Errorf("request == nil")
	}

	var inputVersions []afas.FirmwareVersion
	for _, firmwareVersion := range request.Firmwares {
		if firmwareVersion == nil {
			return nil, fmt.Errorf("firmware version should not be nil")
		}
		inputVersions = append(inputVersions, *firmwareVersion)
	}

	result, err := svc.Controller.CheckFirmwareVersion(
		ctx,
		inputVersions,
	)
	if err != nil {
		return nil, unwrapException(err)
	}
	return &afas.CheckFirmwareVersionResult_{
		ExistStatus: result,
	}, nil
}

func unwrapException(err error) error {
	if err == nil {
		return nil
	}

	var exceptioner thriftExceptioner
	if !errors.As(err, &exceptioner) {
		return err
	}
	return exceptioner.ThriftException()
}