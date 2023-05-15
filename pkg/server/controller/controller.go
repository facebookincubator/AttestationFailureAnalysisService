package controller

import (
	"context"
	"crypto/sha512"
	"fmt"
	"strings"
	"sync"
	"time"

	"lukechampine.com/blake3"

	css_errors "github.com/9elements/converged-security-suite/v2/pkg/errors"
	"github.com/9elements/converged-security-suite/v2/pkg/uefi"
	"github.com/facebookincubator/go-belt/beltctx"
	"github.com/facebookincubator/go-belt/tool/logger"
	fianoUEFI "github.com/linuxboot/fiano/pkg/uefi"

	"github.com/immune-gmbh/AttestationFailureAnalysisService/if/generated/afas"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/if/generated/device"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/analyzers"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/firmwaredb"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/lockmap"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/storage/models"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/types"
)

func init() {
	// We do not modify images, so we can enable an appropriate optimizations.
	fianoUEFI.ReadOnly = true

	// We do not modify images and we are not interested in decompressed data,
	// so we can enable an appropriate optimization.
	fianoUEFI.DisableDecompression = true

	// Ignore "erase polarity conflict" error.
	fianoUEFI.SuppressErasePolarityError = true
}

type noCopy sync.Locker

// Controller implement the high-level logic of the firmware-analysis service.
type Controller struct {
	noCopy noCopy

	Context                   context.Context
	ContextCancel             context.CancelFunc
	FirmwareStorage           Storage
	DeviceGetter              DeviceGetter
	OriginalFWDB              firmwaredb.DB
	OriginalFWImageRepository originalFWImageRepository
	analyzersRegistry         *analyzers.Registry
	analysisDataCalculator    analysisDataCalculatorInterface

	UEFIParseLock *lockmap.LockMap

	closedSignal       chan struct{}
	activeGoroutinesWG sync.WaitGroup
}

func New(
	ctx context.Context,
	firmwareStorage Storage,
	origFirmwareDB firmwaredb.DB,
	origFirmwareRepo originalFWImageRepository,
	analysisDataCalculator analysisDataCalculatorInterface,
	deviceGetter DeviceGetter,
	apiCachePurgeTimeout time.Duration,
) (*Controller, error) {
	ctx = beltctx.WithField(ctx, "module", "controller")

	analyzersRegistry, err := analyzers.NewRegistryWithKnownAnalyzers()
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzers registry: %w", err)
	}

	ctrl := &Controller{
		FirmwareStorage:           firmwareStorage,
		DeviceGetter:              deviceGetter,
		OriginalFWDB:              origFirmwareDB,
		OriginalFWImageRepository: origFirmwareRepo,
		analyzersRegistry:         analyzersRegistry,
		analysisDataCalculator:    analysisDataCalculator,

		UEFIParseLock: lockmap.NewLockMap(),
		closedSignal:  make(chan struct{}),
	}
	ctrl.Context, ctrl.ContextCancel = context.WithCancel(ctx)

	ctrl.launchAsync(ctrl.Context, func(ctx context.Context) {
		ctrl.updateCacheLoop(ctx, apiCachePurgeTimeout)
	})
	return ctrl, nil
}

func (ctrl *Controller) updateCacheLoop(
	ctx context.Context,
	apiCachePurgeTimeout time.Duration,
) {
	apiCachePurgeTicker := time.NewTicker(apiCachePurgeTimeout)
	defer apiCachePurgeTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-apiCachePurgeTicker.C:
			ctrl.purgeAPICache()
		}
	}
}

func (ctrl *Controller) purgeAPICache() {
	ctx := ctrl.Context

	logger.FromCtx(ctx).Infof("purge controller API cache")
	// TODO: purge any cache
}

func uint64deref(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

func getOrigFirmwareInfo(
	ctx context.Context,
	db firmwaredb.DB,
	firmwareVersion string,
	modelID int64,
	cachingPolicy types.CachingPolicy,
) (*firmwaredb.Firmware, error) {
	fws, err := db.Get(ctx, firmwaredb.FilterVersion(firmwareVersion), firmwaredb.FilterModelIDs{modelID})
	if err != nil {
		return nil, err
	}
	if len(fws) != 1 {
		return nil, fmt.Errorf("expected 1 entry, but got %d", len(fws))
	}
	return fws[0], nil
}

// getHostInfo tries to get full information about the host being analyzed.
// If request is being made from the host that is being analyzed, a host can provide information about itself in thrift input structure.
// But that information may not be full
func (ctrl *Controller) getHostInfo(
	ctx context.Context,
	requestHostInfo *afas.HostInfo,
) (*afas.HostInfo, *device.Device) {
	if requestHostInfo == nil {
		return nil, nil
	}

	log := logger.FromCtx(ctx)

	resultHostInfo := *requestHostInfo
	serfDevice := func() *device.Device {
		if resultHostInfo.IsClientHostAnalyzed {
			hostname, _ := ExtractHostnameFromCtx(ctx)
			if len(hostname) > 0 {
				log.Debugf("detected TLS identity hostname: %s", hostname)
				device, err := ctrl.DeviceGetter.GetDeviceByHostname(hostname)
				if err == nil {
					return device
				}
				log.Warnf("failed to get SeRF info for %s: %v", hostname, err)
			}
		}
		if resultHostInfo.AssetID != nil {
			device, err := ctrl.DeviceGetter.GetDeviceByAssetID(*resultHostInfo.AssetID)
			if err == nil {
				return device
			}
			log.Warnf("failed to get SeRF info by asset id %d: %v", *resultHostInfo.AssetID, err)
		}

		if resultHostInfo.Hostname != nil {
			device, err := ctrl.DeviceGetter.GetDeviceByHostname(*resultHostInfo.Hostname)
			if err == nil {
				return device
			}
			log.Warnf("failed to get SeRF info for %s: %v", *resultHostInfo.Hostname, err)
		}
		return nil
	}()
	if serfDevice != nil {
		enrichHostInfo(ctx, serfDevice, false, &resultHostInfo)
	}
	return &resultHostInfo, serfDevice
}

// Close stops the Controller and blocks until all goroutines from launchAsync
// rejoin.
//
// Invariants:
//  1. Close will wait for goroutines to rejoin before invalidating any state
//  2. After Close has been called, launchAsync will fail with context.Canceled
//  3. Goroutines MUST NOT call Close
//  4. Goroutines MUST return promptly when their context is cancelled
func (ctrl *Controller) Close() error {
	ctrl.ContextCancel()
	ctrl.activeGoroutinesWG.Wait()

	err := css_errors.MultiError{}
	err.Add(
		ctrl.FirmwareStorage.Close(),
	)
	return err.ReturnValue()
}

// launchAsync starts the given function in the background. The context passed
// to the function will be cancelled with the call to ctrl.Close(). If the
// controller has already received a call to Close, then this function will
// return the cancellation error (in this case most likely context.Canceled).
//
// Goroutines launched this way MUST NOT call Controller.Close, because it would
// DEADLOCK. See Close for other invariants.
func (ctrl *Controller) launchAsync(ctx context.Context, f func(ctx context.Context)) error {
	// Need to do this first to prevent another thread entering Close() between
	// the `if` and the `go` from returning.
	ctrl.activeGoroutinesWG.Add(1)
	if ctx.Err() != nil {
		ctrl.activeGoroutinesWG.Done()
		return ctx.Err()
	}

	// If another thread calls Close before this goroutine starts, the latter
	// will spin up with an already cancelled context and return promptly, but
	// the Close thread will still wait for it.
	go func() {
		defer ctrl.activeGoroutinesWG.Done()
		f(ctx)
	}()

	return nil
}

func (ctrl *Controller) parseUEFI(b []byte) (*uefi.UEFI, error) {
	// Preventing multiple concurrent parsing of the same image. Compute one and use everywhere, instead
	l := ctrl.UEFIParseLock.Lock(firmwareImageCacheKey(b))
	defer l.Unlock()

	if l.UserData != nil {
		// The value is already computed, just use it.
		return l.UserData.(*uefi.UEFI), nil
	}

	f, err := uefi.ParseUEFIFirmwareBytes(b)
	if err != nil {
		return nil, err
	}

	l.UserData = f
	return f, nil
}

func firmwareMetaCacheKey(imageMeta models.ImageMetadata) string {
	return firmwareCacheKey(imageMeta.HashSHA2_512, imageMeta.HashBlake3_512)
}

func firmwareImageCacheKey(b []byte) string {
	sha512Sum := sha512.Sum512(b)
	blake3Sum := blake3.Sum512(b)
	return firmwareCacheKey(sha512Sum[:], blake3Sum[:])
}

// For security reasons we do two different hashes (to avoid intentional
// collisions).
func firmwareCacheKey(sha512Sum, blake3Sum []byte) string {
	var resultCacheKey strings.Builder
	resultCacheKey.Write(sha512Sum[:])
	resultCacheKey.Write(blake3Sum[:])

	return resultCacheKey.String()
}
