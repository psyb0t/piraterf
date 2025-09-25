package aichteeteapee

type ErrorCode = string

// Error code constants.
const (
	ErrorCodeFileNotFound                 ErrorCode = "FILE_NOT_FOUND"
	ErrorCodeDirectoryListingNotSupported ErrorCode = "DIRECTORY_LISTING_" +
		"NOT_SUPPORTED"
	ErrorCodePathTraversalDenied    ErrorCode = "PATH_TRAVERSAL_DENIED"
	ErrorCodeEndpointNotFound       ErrorCode = "ENDPOINT_NOT_FOUND"
	ErrorCodeMethodNotAllowed       ErrorCode = "METHOD_NOT_ALLOWED"
	ErrorCodeMissingUserID          ErrorCode = "MISSING_USER_ID"
	ErrorCodeInvalidUserID          ErrorCode = "INVALID_USER_ID"
	ErrorCodeValidationFailed       ErrorCode = "VALIDATION_FAILED"
	ErrorCodeBadRequest             ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized           ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden              ErrorCode = "FORBIDDEN"
	ErrorCodeInternalServerError    ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeMissingContentType     ErrorCode = "MISSING_CONTENT_TYPE"
	ErrorCodeUnsupportedContentType ErrorCode = "UNSUPPORTED_CONTENT_TYPE"

	// File upload errors.
	ErrorCodeInvalidMultipartForm ErrorCode = "INVALID_MULTIPART_FORM"
	ErrorCodeNoFileProvided       ErrorCode = "NO_FILE_PROVIDED"
	ErrorCodeFileSaveFailed       ErrorCode = "FILE_SAVE_FAILED"
)
