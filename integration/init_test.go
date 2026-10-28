package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var nativeImageBuildpack string

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	output, err := exec.Command("bash", "-c", "../scripts/package.sh --version 1.2.3").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))

	nativeImageBuildpack, err = filepath.Abs("../build/buildpackage.cnb")
	Expect(err).NotTo(HaveOccurred())

	format.MaxLength = 0
	SetDefaultEventuallyTimeout(60 * time.Second)

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("SpringBootMaven", testSpringBootMaven)
	suite("SpringBootGradle", testSpringBootGradle)
	suite("Quarkus", testQuarkus)
	suite("PublicStaticMain", testPublicStaticMain)

	samplesDir, err := filepath.Abs("./samples")
	Expect(err).To(Succeed())

	if _, err := os.Stat(samplesDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/paketo-buildpacks/samples", samplesDir)
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "failed to clone samples, output:\n%s", out)
	}

	suite.Run(t)
}
