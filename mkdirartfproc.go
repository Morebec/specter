package specter

import (
	"fmt"
	"github.com/morebec/go-errors/errors"
	"os"
	"sync"
)

// DirectoryArtifact is a data structure that can be used by a SpecificationProcessor to directory artifacts
// that can be written by the MakeDirectoryArtifactsProcessor.
type DirectoryArtifact struct {
	Path      string
	FileMode  os.FileMode
	WriteMode WriteMode
}

type MakeDirectoryArtifactsProcessor struct {
	FileSystem FileSystem
}

func (p MakeDirectoryArtifactsProcessor) Name() string {
	return "directory_artifacts_processor"
}

func (p MakeDirectoryArtifactsProcessor) Process(ctx ArtifactProcessingContext) error {
	ctx.Logger.Info("Creating artifact directories ...")

	if err := p.cleanRegistry(ctx); err != nil {
		ctx.Logger.Error("failed cleaning artifact registry")
		return err
	}

	errs := errors.NewGroup(WriteFileArtifactsProcessorErrorCode)
	// create directories concurrently to speed up process.
	var wg sync.WaitGroup
	for _, artifact := range ctx.Artifacts {
		if err := CheckContextDone(ctx); err != nil {
			return err
		}

		dir, ok := artifact.Value.(DirectoryArtifact)
		if !ok {
			continue
		}
		if dir.WriteMode == "" {
			dir.WriteMode = WriteOnceMode
		}

		wg.Add(1)
		go func(ctx ArtifactProcessingContext, artifactName string, dir DirectoryArtifact) {
			defer wg.Done()
			dirPath, err := p.FileSystem.Abs(dir.Path)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed creating directory at %q", dir.Path))
				errs = errs.Append(err)
				return
			}

			dirExists := true
			if _, err := p.FileSystem.StatPath(dirPath); err != nil {
				if !os.IsNotExist(err) {
					ctx.Logger.Error(fmt.Sprintf("failed writing artifact file at %q", dirPath))
					errs = errs.Append(err)
					return
				}
				dirExists = false
			}

			if dir.WriteMode == WriteOnceMode && dirExists {
				ctx.Logger.Info(fmt.Sprintf("Directory %q already exists ... skipping", dirPath))
				return
			}

			ctx.Logger.Info(fmt.Sprintf("Creating directory %q ...", dirPath))
			if err = p.FileSystem.Mkdir(dirPath, dir.FileMode); err != nil {
				ctx.Logger.Error(fmt.Sprintf("failed creating directory at %q", dirPath))
				errs = errs.Append(err)
				return
			}

			if dir.WriteMode != WriteOnceMode {
				ctx.AddToRegistry(artifactName)
			}
		}(ctx, artifact.Name, dir)
	}
	wg.Wait()

	ctx.Logger.Success("Artifact directories created successfully.")

	if errs.HasErrors() {
		return errs
	}

	return nil
}

func (p MakeDirectoryArtifactsProcessor) cleanRegistry(ctx ArtifactProcessingContext) error {
	var wg sync.WaitGroup
	for _, o := range ctx.RegistryArtifacts() {
		if err := CheckContextDone(ctx); err != nil {
			return err
		}

		wg.Add(1)
		go func(ctx ArtifactProcessingContext, o string) {
			defer wg.Done()
			if err := p.FileSystem.Remove(o); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return
				}
				panic(errors.Wrap(err, "failed cleaning artifact registry files"))
			}
			ctx.RemoveFromRegistry(o)
		}(ctx, o)
	}
	wg.Wait()

	return nil
}
