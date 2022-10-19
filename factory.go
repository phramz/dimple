package dimple

var _ Factory = (*factory)(nil)

func WithFn(fn FactoryFn) Factory {
	return &factory{
		fn: fn,
	}
}

func WithContextFn(fn FactoryFnWithContext) Factory {
	return &factory{
		fnWithContext: fn,
	}
}

func WithErrorFn(fn FactoryFnWithError) Factory {
	return &factory{
		fnWithError: fn,
	}
}

type factory struct {
	fn            FactoryFn
	fnWithContext FactoryFnWithContext
	fnWithError   FactoryFnWithError
}

func (f *factory) FactoryFn() FactoryFn {
	return f.fn
}

func (f *factory) FactoryFnWithError() FactoryFnWithError {
	return f.fnWithError
}

func (f *factory) FactoryFnWithContext() FactoryFnWithContext {
	return f.fnWithContext
}
