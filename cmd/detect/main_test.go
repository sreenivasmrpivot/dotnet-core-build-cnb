package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-build-cnb/publish"

	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("app has .csproj", func() {
		when("the app only has a runtime dependency", func() {
			it("it passes", func() {
				Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName.csproj"), []byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>netcoreapp2.2</TargetFramework>
  </PropertyGroup>


  <ItemGroup>
  </ItemGroup>

</Project>`), os.ModePerm)).To(Succeed())
				defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "appName.csproj"))
				code, err := runDetect(factory.Detect)
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
					Provides: []buildplan.Provided{{Name: publish.Publish}},
					Requires: []buildplan.Required{{
						Name:     publish.Publish,
						Metadata: buildplan.Metadata{"build": true},
					}, {
						Name:     "dotnet-sdk",
						Version:  "2.2.0",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					}, {
						Name:     "dotnet-runtime",
						Version:  "2.2.0",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					}},
				}))
			})
		})
		when("the app only has runtime and aspnet dependencies", func() {
			it("it passes", func() {
				Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName.csproj"), []byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>netcoreapp2.2</TargetFramework>
  </PropertyGroup>


  <ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.App" />
    <PackageReference Include="Microsoft.AspNetCore.Razor.Design" Version="2.2.0" PrivateAssets="All" />
  </ItemGroup>

</Project>`), os.ModePerm)).To(Succeed())
				defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "appName.csproj"))
				code, err := runDetect(factory.Detect)
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
					Provides: []buildplan.Provided{{Name: publish.Publish}},
					Requires: []buildplan.Required{{
						Name:     publish.Publish,
						Metadata: buildplan.Metadata{"build": true},
					}, {
						Name:     "dotnet-sdk",
						Version:  "2.2.0",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					}, {
						Name:     "dotnet-runtime",
						Version:  "2.2.0",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					}, {
						Name:     "dotnet-aspnet",
						Version:  "2.2.0",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					}},
				}))
			})
		})
	})

	when("app has .fsproj", func() {
		it("it passes", func() {
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName.fsproj"), []byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>netcoreapp2.2</TargetFramework>
  </PropertyGroup>


  <ItemGroup>
  </ItemGroup>

</Project>`), os.ModePerm)).To(Succeed())
			defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "appName.fsproj"))
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))

		})
	})

	when("app has multiple proj files", func(){
		var projBody []byte
		projBody = []byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>netcoreapp2.2</TargetFramework>
  </PropertyGroup>


  <ItemGroup>
  </ItemGroup>

</Project>`)

		it(" that are the same type it takes the first proj file found", func() {
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName.csproj"), projBody, os.ModePerm)).To(Succeed())
			defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "appName.csproj"))
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "another.csproj"), projBody, os.ModePerm)).To(Succeed())
			defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "another.csproj"))
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
		})

		it(" that are the differnt types it fails", func() {
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName.csproj"), projBody, os.ModePerm)).To(Succeed())
			defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "appName.csproj"))
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "another.vbproj"), projBody, os.ModePerm)).To(Succeed())
			defer os.RemoveAll(filepath.Join(factory.Detect.Application.Root, "another.fsproj"))
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
		})
	})

	when("app has no proj file", func() {
		it("it fails", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("no proj file found")))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})
}
