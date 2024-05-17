package errors_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	cerr "github.com/aserto-dev/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound      = newErr("E10001", codes.NotFound, http.StatusNotFound, "not found")
	ErrAlreadyExists = newErr("E10002", codes.AlreadyExists, http.StatusConflict, "already exists")
)

func newErr(code string, statusCode codes.Code, httpCode int, msg string) *cerr.AsertoError {
	return cerr.NewAsertoError(code, statusCode, httpCode, msg)
}

func TestDoubleCerr(t *testing.T) {
	assert := require.New(t)

	err := ErrNotFound.Err(ErrAlreadyExists)

	assert.Contains(err.Error(), "not found")
	assert.Contains(err.Error(), "already exists")
}

func TestDoubleCerrWithMsg(t *testing.T) {
	assert := require.New(t)

	err := ErrNotFound.Err(ErrAlreadyExists).Msg("failed to setup")

	assert.Contains(err.Error(), "not found")
	assert.Contains(err.Error(), "already exists")
}

func TestWithEmptyMsg(t *testing.T) {
	assert := require.New(t)

	err := ErrNotFound.Msg("")

	fields := err.Fields()
	assert.Nil(fields["msg"])

	err = ErrNotFound.Msg("bla")

	fields = err.Fields()
	assert.NotNil(fields["msg"])
}

func TestError(t *testing.T) {
	assert := require.New(t)

	err := ErrNotFound.Msg("bla").Err(errors.New("boom"))
	err2 := ErrNotFound.Msg("bla").Msg("ala")
	err3 := ErrNotFound.Err(errors.New("boom")).Msg("bla").Msg("ala")
	err4 := ErrNotFound.Err(errors.New("boom")).Err(errors.New("pow")).Msg("bla").Msg("ala")
	err5 := ErrNotFound.Err(errors.New("boom"))
	err6 := ErrNotFound.Err(errors.New("boom")).Err(errors.New("pow"))
	err7 := ErrNotFound.Msg("bla")

	assert.Equal(err.Error(), "E10001 not found: boom: bla")
	assert.Equal(err2.Error(), "E10001 not found: bla: ala")
	assert.Equal(err3.Error(), "E10001 not found: boom: bla: ala")
	assert.Equal(err4.Error(), "E10001 not found: boom: pow: bla: ala")
	assert.Equal(err5.Error(), "E10001 not found: boom")
	assert.Equal(err6.Error(), "E10001 not found: boom: pow")
	assert.Equal(err7.Error(), "E10001 not found: bla")
}

func TestWithGrpcStatusCode(t *testing.T) {
	assert := require.New(t)
	err := ErrNotFound.WithGRPCStatus(codes.Canceled)
	assert.Equal(err.StatusCode, codes.Canceled)
}

func TestWithHttpStatusCode(t *testing.T) {
	assert := require.New(t)
	err := ErrNotFound.WithHTTPStatus(http.StatusAccepted)
	assert.Equal(err.HTTPCode, http.StatusAccepted)
}

func TestFromGRPCStatus(t *testing.T) {
	assert := require.New(t)

	initialErr := ErrNotFound
	initialErr = initialErr.Str("email", "testuser@mail.com").Msg("foo")

	grpcStatus := status.New(initialErr.StatusCode, initialErr.Error())
	grpcStatus, err := grpcStatus.WithDetails(&errdetails.ErrorInfo{
		Reason:   "1234",
		Metadata: initialErr.Data(),
		Domain:   initialErr.Code,
	})
	if err != nil {
		assert.Fail(err.Error())
	}

	transformedErr := cerr.FromGRPCStatus(*grpcStatus)

	assert.True(initialErr.SameAs(transformedErr))

	assert.Equal(initialErr.Error(), transformedErr.Error())
	assert.Equal(initialErr.Message, transformedErr.Message)
}

func TestUnwrapNilErr(t *testing.T) {
	assert := require.New(t)

	err := cerr.UnwrapAsertoError(nil)

	assert.Nil(err)
}

func TestEquals(t *testing.T) {
	assert := require.New(t)

	err1 := ErrAlreadyExists.Msgf("error 1").Str("key1", "val1").Err(errors.New("boom"))
	err2 := ErrAlreadyExists.Msgf("error 2").Str("key2", "val2").Err(errors.New("zoom"))

	assert.True(cerr.Equals(err1, err2))
}

func TestEqualsNil(t *testing.T) {
	assert := require.New(t)

	assert.True(cerr.Equals(nil, nil))
}

func TestEqualsOneNil(t *testing.T) {
	assert := require.New(t)

	assert.False(cerr.Equals(ErrNotFound, nil))
}

func TestEqualsNormalErrorOneNil(t *testing.T) {
	assert := require.New(t)

	assert.False(cerr.Equals(errors.New("boom"), nil))
}

func TestEqualsErrCerr(t *testing.T) {
	assert := require.New(t)

	assert.False(cerr.Equals(errors.New("boom"), ErrNotFound))
}

func TestEqualsFalse(t *testing.T) {
	assert := require.New(t)

	assert.False(cerr.Equals(ErrAlreadyExists, ErrNotFound))
}

func TestEqualsNormalErrors(t *testing.T) {
	assert := require.New(t)

	assert.False(cerr.Equals(errors.New("boom1"), errors.New("boom2")))
}

func TestCodeToAsertoError(t *testing.T) {
	assert := require.New(t)

	asertoErr := cerr.CodeToAsertoError("E10001")

	assert.NotNil(asertoErr)
	assert.True(cerr.Equals(asertoErr, ErrNotFound))
}

func TestCodeToAsertoErrorInvalidCode(t *testing.T) {
	assert := require.New(t)

	asertoErr := cerr.CodeToAsertoError("E20009")

	assert.Nil(asertoErr)
}

func TestWithGrpcError(t *testing.T) {
	assert := require.New(t)
	aerr := cerr.NewAsertoError("E000001", codes.Unavailable, http.StatusServiceUnavailable, "failed to setup").WithGRPCStatus(codes.Aborted)
	berr := errors.Wrap(aerr, "new err")

	unAerr := cerr.UnwrapAsertoError(berr)
	assert.Equal(codes.Aborted, unAerr.GRPCStatus().Code())
}

func TestWithHttpError(t *testing.T) {
	assert := require.New(t)
	aerr := cerr.NewAsertoError("E000001", codes.Unavailable, http.StatusServiceUnavailable, "failed to setup").WithHTTPStatus(http.StatusNotAcceptable)

	unAerr := cerr.UnwrapAsertoError(aerr)
	assert.Equal(http.StatusNotAcceptable, unAerr.HTTPCode)
}

// returns nil logger if error is nil.
func TestLoggerWithNilError(t *testing.T) {
	assert := require.New(t)

	var err error
	logger := cerr.Logger(err)
	assert.Nil(logger)
}

func TestLoggerWithWrappedNilError(t *testing.T) {
	assert := require.New(t)

	var err error
	ctx := context.Background()

	logger := cerr.Logger(cerr.WithContext(err, ctx))
	assert.Nil(logger)
}

func TestLoggerWithWrappedErrorsWithEmptyContext(t *testing.T) {
	assert := require.New(t)

	ctx := context.Background()
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx)
	wrappedErr := errors.Wrap(err, "wrapped error")

	logger := cerr.Logger(wrappedErr)
	assert.Nil(logger)
}

func TestLoggerWithWrappedErrorsWithLoggerContext(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)

	ctx := context.Background()
	ctx = initialLogger.WithContext(ctx)
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx)
	wrappedErr := errors.Wrap(err, "wrapped error")

	logger := cerr.Logger(wrappedErr)
	assert.NotNil(logger)
	assert.Equal(logger, zerolog.Ctx(ctx))
}

func TestLoggerWithWrappedMultipleWithoutErrorsWithContext(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)

	ctx := context.Background()
	ctx = initialLogger.WithContext(ctx)
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx)
	errWithoutCtx := cerr.NewAsertoError("E00002", codes.Internal, http.StatusInternalServerError, "internal error")
	wrappedErr := errWithoutCtx.Err(errors.Wrap(err, "wrapped error"))

	logger := cerr.Logger(wrappedErr)
	assert.NotNil(logger)
	assert.Equal(logger, zerolog.Ctx(ctx))
}

func TestLoggerWithWrappedMultipleErrorsWithContext(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)

	ctx := context.Background()
	ctx = initialLogger.WithContext(ctx)
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx)
	errWithoutCtx := cerr.NewAsertoError("E00002", codes.Internal, http.StatusInternalServerError, "internal error")
	wrappedErr := errors.Wrap(errWithoutCtx.Err(err), "wrapped error")

	logger := cerr.Logger(wrappedErr)
	assert.NotNil(logger)
	assert.Equal(logger, zerolog.Ctx(ctx))
}

func TestLoggerWithWrappedMultipleErrorsWithMultipleContexts(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)
	ctx1 := context.Background()
	ctx2 := initialLogger.WithContext(ctx1)
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx1)
	wrappedErr := cerr.WithContext(cerr.WithContext(err, ctx2), ctx1)

	logger := cerr.Logger(wrappedErr)
	ctx1Logger := zerolog.Ctx(ctx1)
	ctx2Logger := zerolog.Ctx(ctx2)

	assert.NotNil(logger)
	assert.NotEqual(logger, ctx1Logger)
	assert.Equal(logger, ctx2Logger)
}

func TestLoggerWithWrappedMultipleErrorsWithMultipleContextsOuter(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)
	ctx1 := context.Background()
	ctx2 := initialLogger.WithContext(ctx1)
	err := cerr.WithContext(cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error"), ctx1)
	err2 := cerr.WithContext(cerr.NewAsertoError("E00002", codes.Internal, http.StatusInternalServerError, "internal error"), ctx2)
	wrappedErr := errors.Wrap(errors.Wrap(err2, err.Error()), "wrapped error")

	logger := cerr.Logger(wrappedErr)
	ctx1Logger := zerolog.Ctx(ctx1)
	ctx2Logger := zerolog.Ctx(ctx2)

	assert.NotNil(logger)
	assert.NotEqual(logger, ctx1Logger)
	assert.Equal(logger, ctx2Logger)
}

func TestLoggerWithWrappedMultipleAsertoErrorsWithMultipleContextsOuter(t *testing.T) {
	assert := require.New(t)
	initialLogger := zerolog.New(os.Stderr)
	ctx1 := context.Background()
	ctx2 := initialLogger.WithContext(ctx1)
	err := cerr.NewAsertoError("E00001", codes.Internal, http.StatusInternalServerError, "internal error").Ctx(ctx1)
	err2 := cerr.NewAsertoError("E00002", codes.Internal, http.StatusInternalServerError, "internal error").Ctx(ctx2)
	wrappedErr := errors.Wrap(errors.Wrap(err2, err.Error()), "wrapped error")

	logger := cerr.Logger(wrappedErr)
	ctx1Logger := zerolog.Ctx(ctx1)
	ctx2Logger := zerolog.Ctx(ctx2)

	assert.NotNil(logger)
	assert.NotEqual(logger, ctx1Logger)
	assert.Equal(logger, ctx2Logger)
}
