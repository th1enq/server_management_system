package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/th1enq/server_management_system/internal/domain/scope"
)

func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("valid_scope", func(fl validator.FieldLevel) bool {
			scopes, ok := fl.Field().Interface().([]scope.APIScope)
			if !ok {
				return false
			}
			for _, s := range scopes {
				if !scope.IsValidScope(s) {
					return false
				}
			}
			return true
		})
	}

}
