package diskchk

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskUsageConfig is used for configuring the go-diskusage check.
//
// "Path" is _required_; path to check directory/drive (ex. /home/user)
// "WarningThreshold" is _required_; set percent (more than 0 and less 100) of free space at specified path,
//
//	which triggers warning.
//
// "CriticalThreshold" is _required_; set percent (more than 0 and less 100) of free space at specified path,
//
//	which triggers critical.
type DiskUsageConfig struct {
	Path              string
	WarningThreshold  float64
	CriticalThreshold float64
}

// DiskUsage implements the "ICheckable" interface.
type DiskUsage struct {
	Config *DiskUsageConfig
}

func NewDiskUsage(cfg *DiskUsageConfig) (*DiskUsage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Passed in config cannot be nil")
	}

	if err := validateDiskUsageConfig(cfg); err != nil {
		return nil, fmt.Errorf("Unable to validate diskusage config: %v", err)
	}

	return &DiskUsage{
		Config: cfg,
	}, nil
}

// Status is used for performing a diskusage check against a dependency; it satisfies
// the "ICheckable" interface.
func (d *DiskUsage) Status(ctx context.Context) (interface{}, error) {
	stats, err := disk.Usage(d.Config.Path)

	if err != nil {
		return nil, fmt.Errorf("Error getting disk usage: %v", err)
	}

	diskUsage := stats.UsedPercent

	if diskUsage >= d.Config.CriticalThreshold {
		return nil, fmt.Errorf("Critical: disk usage too high %.2f percent", diskUsage)
	}

	if diskUsage >= d.Config.WarningThreshold {
		return nil, fmt.Errorf("Warning: disk usage too high %.2f percent", diskUsage)
	}

	return nil, nil
}

func validateDiskUsageConfig(cfg *DiskUsageConfig) error {
	if cfg == nil {
		return fmt.Errorf("Main config cannot be nil")
	}

	if cfg.Path == "" {
		return fmt.Errorf("Path can not be nil")
	}

	if cfg.WarningThreshold > 100.0 || cfg.WarningThreshold < 0 {
		return fmt.Errorf("Invalid warning threshold value (more 100 or less 0)")
	}

	if cfg.CriticalThreshold > 100.0 || cfg.CriticalThreshold < 0 {
		return fmt.Errorf("Invalid critical threshold value (more 100 or less 0)")
	}

	return nil
}
