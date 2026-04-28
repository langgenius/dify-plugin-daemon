package slim

import (
	"encoding/json"
	"io"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

const OutputJSON = "json"

type ExtractOptions struct {
	PluginID string
	Action   string
	Path     string
	Output   string
}

type ExtractResult struct {
	UniqueIdentifier string                            `json:"unique_identifier,omitempty"`
	Manifest         plugin_entities.PluginDeclaration `json:"manifest"`
	Verification     *decoder.Verification             `json:"verification,omitempty"`
}

type ExtractArgs struct {
	TenantID string `json:"tenant_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Data     any    `json:"data"`
}

func RunExtract(cfg *SlimConfig, opts ExtractOptions, w io.Writer) error {
	if opts.Output == "" {
		opts.Output = OutputJSON
	}
	if opts.Output != OutputJSON {
		return NewError(ErrInvalidInput, "unsupported output format: "+opts.Output)
	}

	var result *ExtractResult
	var err error
	switch cfg.Mode {
	case ModeLocal:
		result, err = ExtractLocal(opts, &cfg.Local)
	case ModeRemote:
		result, err = ExtractRemote(opts, &cfg.Remote)
	default:
		return NewError(ErrUnknownMode, cfg.Mode)
	}
	if err != nil {
		return err
	}

	args := ExtractArgs{
		TenantID: "00000000-0000-0000-0000-000000000000",
		Data:     BuildExtractData(*result),
	}
	if opts.Action != "" {
		payload, err := BuildActionArgs(result.Manifest, opts.Action)
		if err != nil {
			return err
		}
		args.Data = payload
	}

	if err := json.NewEncoder(w).Encode(args); err != nil {
		return NewError(ErrInvalidInput, "failed to encode extract result: "+err.Error())
	}
	return nil
}
