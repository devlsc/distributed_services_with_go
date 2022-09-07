package server

import (
    "testing"

	api "github.com/devlsc/distributed_services_with_go/prolog/api/v1"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.LogClient,
		config *Config,
	){
		"produce/consumee a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                     testProduceConsumeStream,
		"consume past log boundary fails":                     testConsumePastBoundary,
	} {
		t.Run(scenario, func(t *testing.T) {
            client, config, teardown := setupTest(t, nil)
            defer teardown()
            fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config))(
    client api.LogClient,
    cfg *Config,
    teardown func(),
){
    t.Helper()
    return nil, nil, nil
}
