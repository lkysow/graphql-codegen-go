package pkg

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/lkysow/graphql-codegen-go/internal"
)

func Generate(schemaFile string, pkgName string, outputFile string) error {
	if schemaFile == "" {
		return fmt.Errorf("schemaFile not defined")
	}
	if pkgName == "" {
		return fmt.Errorf("packageName not defined")
	}
	if outputFile == "" {
		return fmt.Errorf("outputFile not defined")
	}

	config := internal.Config{
		Schemas: []string{schemaFile},
		Outputs: []internal.OutputItem{
			{OutputPath: outputFile, PackageName: pkgName},
		},
	}

	// Combine all schemas.
	inputSchemas, err := internal.ReadSchemas(config.Schemas)
	if err != nil {
		return err
	}

	for _, o := range config.Outputs {
		output, err := internal.NewFileOutput(o.OutputPath)
		if err != nil {
			return errors.Wrapf(err, "failed to create output to %s", o.OutputPath)
		}

		loadedDocs, err := internal.LoadSchemas(inputSchemas)
		if err != nil {
			return errors.Wrapf(err, "failed to parse input schemas")
		}

		gen := internal.NewGoGenerator(output, o.Entities, o.PackageName)
		if err := gen.Generate(loadedDocs); err != nil {
			return errors.Wrapf(err, "failed to generate go structs")
		}

		if err := output.Close(); err != nil {
			return err
		}
	}
	return nil
}
