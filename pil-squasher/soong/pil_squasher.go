//
// SPDX-FileCopyrightText: The LineageOS Project
// SPDX-License-Identifier: Apache-2.0
//

package pil_squasher

import (
	"github.com/google/blueprint/proptools"

	"android/soong/android"
)

var (
	pctx = android.NewPackageContext("android/soong/pil_squasher")
)

func init() {
	android.RegisterModuleType("pil_squasher", pilSquasherFactory)
	pctx.Import("android/soong/android")
}

type pilSquasherProperties struct {
	// The MDT header file
	Mdt *string `android:"path"`

	// The chunks (.b00, .b01, etc.).
	Srcs []string `android:"path"`

	// Output file name. Defaults to {name}.mbn
	Stem *string

	// Install to a subdirectory of the install path
	Relative_install_path *string
}

type pilSquasher struct {
	android.ModuleBase
	properties pilSquasherProperties

	outputFile  android.WritablePath
	installPath android.InstallPath
}

func pilSquasherFactory() android.Module {
	module := &pilSquasher{}
	module.AddProperties(&module.properties)
	android.InitAndroidArchModule(module, android.DeviceSupported, android.MultilibCommon)
	return module
}

func (p *pilSquasher) inVendorRamdisk() bool {
	return p.ModuleBase.InVendorRamdisk() || p.ModuleBase.InstallInVendorRamdisk()
}

func (p *pilSquasher) onlyInVendorRamdisk() bool {
	return p.ModuleBase.InstallInVendorRamdisk()
}

func (p *pilSquasher) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	mdtPath := android.PathForModuleSrc(ctx, proptools.String(p.properties.Mdt))
	srcPaths := android.PathsForModuleSrc(ctx, p.properties.Srcs)

	stem := proptools.StringDefault(p.properties.Stem, ctx.ModuleName()+".mbn")
	p.outputFile = android.PathForModuleOut(ctx, stem)

	rule := android.NewRuleBuilder(pctx, ctx)

	cmd := rule.Command().BuiltTool("pil-squasher")
	cmd.Output(p.outputFile)
	cmd.Input(mdtPath)
	cmd.Implicits(srcPaths)

	rule.Build("pil_squasher", "Squashing firmware: "+p.outputFile.Base())

	p.installPath = android.PathForModuleInstall(ctx, "firmware")

	if subdir := proptools.String(p.properties.Relative_install_path); subdir != "" {
		p.installPath = p.installPath.Join(ctx, subdir)
	}

	ctx.InstallFile(p.installPath, p.outputFile.Base(), p.outputFile)

	ctx.SetOutputFiles(android.Paths{p.outputFile}, "")
}

func (p *pilSquasher) AndroidMkEntries() []android.AndroidMkEntries {
	if p.outputFile == nil {
		return nil
	}

	nameSuffix := ""
	if p.inVendorRamdisk() && !p.onlyInVendorRamdisk() {
		nameSuffix = ".vendor_ramdisk"
	}

	return []android.AndroidMkEntries{
		{
			Class:      "ETC",
			SubName:    nameSuffix,
			OutputFile: android.OptionalPathForPath(p.outputFile),
			ExtraEntries: []android.AndroidMkExtraEntriesFunc{
				func(ctx android.AndroidMkExtraEntriesContext, entries *android.AndroidMkEntries) {
					entries.SetPath("LOCAL_MODULE_PATH", p.installPath)
					entries.SetString("LOCAL_INSTALLED_MODULE_STEM", p.outputFile.Base())
				},
			},
		},
	}
}
