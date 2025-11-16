package handler

import (
	"fmt"
	"runtime"

	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"
	validation "MgApplication/api-validation"

	"github.com/gin-gonic/gin"
)

func BindAndValidate(ctx *gin.Context, req interface{}, bindUri, bindQuery, bindJson bool) error {

	// Get the handler's file and line number
	_, file, line, _ := runtime.Caller(1)
	callerInfo := fmt.Sprintf("%s:%d", file, line)

	if bindUri {
		if err := ctx.ShouldBindUri(req); err != nil {
			log.Error(ctx, "Calling function: %s, Failed to bind URI parameters: %s", callerInfo, err.Error())
			apierrors.HandleBindingError(ctx, err)
			return err
		}
	}

	if bindQuery {
		if err := ctx.ShouldBindQuery(req); err != nil {
			log.Error(ctx, "Calling function: %s, Failed to bind query parameters: %s", callerInfo, err.Error())
			apierrors.HandleBindingError(ctx, err)
			return err
		}
	}

	if bindJson {
		if err := ctx.ShouldBindJSON(req); err != nil {
			log.Error(ctx, "Calling function: %s, Failed to bind JSON parameters: %s", callerInfo, err.Error())
			apierrors.HandleBindingError(ctx, err)
			return err
		}
	}

	// Validate the struct
	if err := validation.ValidateStruct(req); err != nil {
		log.Error(ctx, "Calling function: %s, Validation error: %s", callerInfo, err.Error())
		apierrors.HandleValidationError(ctx, err)
		return err
	}

	return nil
} //@API team, please use this helper function to refactor the commonly used code
