package example

import (
	"os"
	"testing"

	"github.com/coopnorge/member-lib/configloader"
	"github.com/stretchr/testify/assert"
)

func Test_LoadIP(t *testing.T) {
	os.Clearenv()
	t.Setenv("ECOM_ENVIRONMENT", "production")
	t.Setenv("ECOM_DEBUG", "true")

	t.Setenv("ECOM_DATABASE_HOST", "localhost")
	t.Setenv("ECOM_DATABASE_DB_USER", "Alfredo")
	t.Setenv("ECOM_DATABASE_PORT", "5432")
	t.Setenv("ECOM_DATABASE_PASSWORD", "secret")

	t.Setenv("ECOM_TIMEOUT", "5s")
	t.Setenv("ECOM_SOME_IP", "10.0.0.3")
	t.Setenv("ECOM_ALLOWED_IPS", "192.168.1.1,192.168.1.2")

	var cfg Config
	err := configloader.Load(
		&cfg,
		configloader.WithPrefix("ECOM"),
		configloader.WithNameTag("env"),
		configloader.WithTypeHandler(IneedAdifferentWayToLoadIps),
	)
	assert.NoError(t, err)

}
