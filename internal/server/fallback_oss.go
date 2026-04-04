package server

import (
	"github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type FallbackOSS struct {
	primary  oss.OSS
	fallback oss.OSS
}

func NewFallbackOSS(primary oss.OSS, fallback oss.OSS) *FallbackOSS {
	return &FallbackOSS{
		primary:  primary,
		fallback: fallback,
	}
}

func (f *FallbackOSS) Save(key string, data []byte) error {
	primaryErr := f.primary.Save(key, data)
	if primaryErr != nil {
		// primary failed, save to fallback synchronously to ensure data is available
		log.Warn("fallback: primary Save failed, saving to local fallback", "key", key, "error", primaryErr)
		if fallbackErr := f.fallback.Save(key, data); fallbackErr != nil {
			log.Error("fallback: both primary and local Save failed", "key", key, "primaryError", primaryErr, "fallbackError", fallbackErr)
			return primaryErr
		}
		return nil
	}
	return nil
}

func (f *FallbackOSS) Load(key string) ([]byte, error) {
	data, err := f.primary.Load(key)
	if err == nil {
		return data, nil
	}

	fallbackData, fallbackErr := f.fallback.Load(key)
	if fallbackErr != nil {
		return nil, err
	}

	log.Warn("fallback: primary Load failed, using local cache", "key", key, "error", err)
	return fallbackData, nil
}

func (f *FallbackOSS) Exists(key string) (bool, error) {
	exists, err := f.primary.Exists(key)
	if err == nil {
		return exists, nil
	}

	fallbackExists, fallbackErr := f.fallback.Exists(key)
	if fallbackErr != nil {
		return false, err
	}

	log.Warn("fallback: primary Exists failed, using local cache", "key", key, "error", err)
	return fallbackExists, nil
}

func (f *FallbackOSS) State(key string) (oss.OSSState, error) {
	state, err := f.primary.State(key)
	if err == nil {
		return state, nil
	}

	fallbackState, fallbackErr := f.fallback.State(key)
	if fallbackErr != nil {
		return oss.OSSState{}, err
	}

	log.Warn("fallback: primary State failed, using local cache", "key", key, "error", err)
	return fallbackState, nil
}

func (f *FallbackOSS) List(prefix string) ([]oss.OSSPath, error) {
	paths, err := f.primary.List(prefix)
	if err == nil {
		return paths, nil
	}

	fallbackPaths, fallbackErr := f.fallback.List(prefix)
	if fallbackErr != nil {
		return nil, err
	}

	log.Warn("fallback: primary List failed, using local cache", "prefix", prefix, "error", err)
	return fallbackPaths, nil
}

func (f *FallbackOSS) Delete(key string) error {
	return f.primary.Delete(key)
}

func (f *FallbackOSS) Type() string {
	return f.primary.Type()
}
