package diskchk

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewDiskUsage(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path and set thresholds", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  5,
			CriticalThreshold: 5,
		}

		du, err := NewDiskUsage(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(du).ToNot(BeNil())

	})

	t.Run("Should error with a nil cfg", func(t *testing.T) {
		du, err := NewDiskUsage(nil)

		Expect(du).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("Passed in config cannot be nil"))
	})

	t.Run("Bad config should error", func(t *testing.T) {
		du, err := NewDiskUsage(&DiskUsageConfig{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unable to validate diskusage config"))
		Expect(du).To(BeNil())
	})
}

func TestValidateDiskUsageConfig(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error with nil main config", func(t *testing.T) {
		var cfg *DiskUsageConfig
		err := validateDiskUsageConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Main config cannot be nil"))
	})

	t.Run("Config must have path set", func(t *testing.T) {
		cfg := &DiskUsageConfig{}

		err := validateDiskUsageConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Path can not be nil"))
	})

	t.Run("Should error if warning threshold value set out of bounds", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  -1,
			CriticalThreshold: 100,
		}

		err := validateDiskUsageConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Invalid warning threshold value"))
	})

	t.Run("Should error if critical threshold value set out of bounds", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  10,
			CriticalThreshold: 101,
		}

		err := validateDiskUsageConfig(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Invalid critical threshold value"))
	})

}

func TestDiskUsageStatus(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Should error when critical threshold reached", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              "/unknown/path",
			WarningThreshold:  50.0,
			CriticalThreshold: 50.0,
		}

		du, err := NewDiskUsage(cfg)
		if err != nil {
			t.Fatal(err)
		}

		_, err = du.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Error getting disk usage"))
	})

	t.Run("Should error when critical threshold reached", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  90,
			CriticalThreshold: 1,
		}

		du, err := NewDiskUsage(cfg)
		if err != nil {
			t.Fatal(err)
		}

		_, err = du.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Critical: disk usage too high"))
	})

	t.Run("Should error when warning threshold reached", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  1,
			CriticalThreshold: 99,
		}

		du, err := NewDiskUsage(cfg)
		if err != nil {
			t.Fatal(err)
		}
		_, err = du.Status()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Warning: disk usage too high"))
	})

	t.Run("Shouldn't return error when everything is ok", func(t *testing.T) {
		cfg := &DiskUsageConfig{
			Path:              os.TempDir(),
			WarningThreshold:  99,
			CriticalThreshold: 99,
		}

		du, err := NewDiskUsage(cfg)
		if err != nil {
			t.Fatal(err)
		}
		_, err = du.Status()
		Expect(err).To(BeNil())
	})
}
