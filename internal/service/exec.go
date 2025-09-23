package service

import "errors"

var (
	ErrUnauthorizedLanggenius = errors.New(`
		plugin installation blocked: this plugin claims to be from Langgenius but lacks official signature verification. 
		Set DISABLE_UNAUTHORIZED_LANGGENIUS_PACKAGE=false to allow installation (not recommended)`,
	)
)
