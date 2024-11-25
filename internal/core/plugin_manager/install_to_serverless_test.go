package plugin_manager

import (
	"testing"
)

func TestExtractVariables(t *testing.T) {
	message := "endpoint=${endpoint},name=${name},id=${id}"
	variables := extractVariables(message)
	if variables["endpoint"] != "${endpoint}" {
		t.Errorf("endpoint is not correct")
	}
	if variables["name"] != "${name}" {
		t.Errorf("name is not correct")
	}
	if variables["id"] != "${id}" {
		t.Errorf("id is not correct")
	}
}
