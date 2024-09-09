package specter

type UnitPreprocessor interface {
	Preprocess(PipelineContext, []Unit) ([]Unit, error)
	Name() string
}

type UnitPreprocessorAdapter struct {
	PreprocessFunc func(PipelineContext, []Unit) ([]Unit, error)
	name           string
}

func (u UnitPreprocessorAdapter) Preprocess(ctx PipelineContext, units []Unit) ([]Unit, error) {
	if u.PreprocessFunc == nil {
		return units, nil
	}

	return u.PreprocessFunc(ctx, units)
}

func (u UnitPreprocessorAdapter) Name() string {
	return u.name
}

func UnitPreprocessorFunc(name string, processFunc func(PipelineContext, []Unit) ([]Unit, error)) *UnitPreprocessorAdapter {
	return &UnitPreprocessorAdapter{PreprocessFunc: processFunc, name: name}
}
