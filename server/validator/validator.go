package validator

import libvalidator "github.com/go-playground/validator/v10"

type Validator struct {
	inner *libvalidator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		inner: libvalidator.New(),
	}
}

func (v *Validator) Validate(i interface{}) error {
	return v.inner.Struct(i)
}
