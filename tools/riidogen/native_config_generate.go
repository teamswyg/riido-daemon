package main

func generateNativeConfigPlan(specPath, templatePath, outPath string) error {
	spec, err := loadNativeConfigPlan(specPath)
	if err != nil {
		return err
	}
	rendered, err := renderSpec(spec, templatePath)
	if err != nil {
		return err
	}
	return writeGeneratedFile(outPath, rendered)
}
