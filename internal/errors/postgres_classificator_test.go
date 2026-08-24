package errors

import (
	stdErrors "errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPostgresErrorClassifier(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	assert.Equal(t, NonRetriable, classifier.Classify(nil))
	assert.Equal(t, NonRetriable, classifier.Classify(stdErrors.New("ordinary error")))
	assert.Equal(t, Retriable, classifier.Classify(&pgconn.PgError{Code: pgerrcode.ConnectionFailure}))
	assert.Equal(t, NonRetriable, classifier.Classify(&pgconn.PgError{Code: pgerrcode.UniqueViolation}))
	assert.Equal(t, NonRetriable, classifier.Classify(&pgconn.PgError{Code: "99999"}))

	assert.Equal(t, Retriable, СlassifyPgError(&pgconn.PgError{Code: pgerrcode.SerializationFailure}))
	assert.Equal(t, Retriable, СlassifyPgError(&pgconn.PgError{Code: pgerrcode.CannotConnectNow}))
	assert.Equal(t, NonRetriable, СlassifyPgError(&pgconn.PgError{Code: pgerrcode.SyntaxError}))
}
