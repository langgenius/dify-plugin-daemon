package service

import "errors"

var (
	ErrUnauthorizedLanggenius = errors.New(`
		plugin installation blocked: this plugin claims to be from Langgenius but lacks official signature verification. 
		Set ENFORCE_LANGGENIUS_PLUGIN_SIGNATURES=false to allow installation (not recommended)`,
	)
)
