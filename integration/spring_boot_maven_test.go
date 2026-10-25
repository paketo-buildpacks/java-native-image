package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"
)

func testSpringBootMaven(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithNoColor()
		docker = occam.NewDocker()
	})

	context("from source", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds and runs a Spring Boot native image with Maven", func() {
			var err error
			source, err = occam.Source(filepath.Join("samples", "java", "native-image", "spring-boot-native-image-maven"))
			Expect(err).NotTo(HaveOccurred())

			var buildLogs fmt.Stringer
			image, buildLogs, err = pack.Build.
				WithBuildpacks(nativeImageBuildpack).
				WithEnv(map[string]string{
					"BP_NATIVE_IMAGE":          "true",
					"BP_MAVEN_BUILD_ARGUMENTS": "-Dmaven.test.skip=true --no-transfer-progress -Pnative package",
					"BP_JVM_VERSION":           "25",
				}).
				WithTrustBuilder().
				WithPullPolicy("if-not-present").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), buildLogs.String())

			Expect(buildLogs).To(ContainLines(ContainSubstring("Paketo Buildpack for BellSoft Liberica")))
			Expect(buildLogs).To(ContainLines(ContainSubstring("Paketo Buildpack for Maven")))
			Expect(buildLogs).To(ContainLines(ContainSubstring("Paketo Buildpack for Executable JAR")))
			Expect(buildLogs).To(ContainLines(ContainSubstring("Paketo Buildpack for Spring Boot")))
			Expect(buildLogs).To(ContainLines(ContainSubstring("Paketo Buildpack for Native Image")))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring("UP")).OnPort(8080).WithEndpoint("/actuator/health"))
		})
	})
}
