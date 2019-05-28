package http_

import (
	"encoding/base64"
	"errors"
	"github.com/bmizerany/pat"
	"github.com/searKing/golib/x/log"
	"github.com/tus/tusd"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const UploadLengthDeferred = "1"

var (
	reExtractFileID  = regexp.MustCompile(`([^/]+)\/?$`)
	reForwardedHost  = regexp.MustCompile(`host=([^,]+)`)
	reForwardedProto = regexp.MustCompile(`proto=(https?)`)
	reMimeType       = regexp.MustCompile(`^[a-z]+\/[a-z\-\+]+$`)
)

// HTTPError represents an error with an additional status code attached
// which may be used when this error is sent in a HTTP response.
// See the net/http package for standardized status codes.
type HTTPError interface {
	error
	StatusCode() int
	Body() []byte
}

type httpError struct {
	error
	statusCode int
}

func (err httpError) StatusCode() int {
	return err.statusCode
}

func (err httpError) Body() []byte {
	return []byte(err.Error())
}

// NewHTTPError adds the given status code to the provided error and returns
// the new error instance. The status code may be used in corresponding HTTP
// responses. See the net/http package for standardized status codes.
func NewHTTPError(err error, statusCode int) HTTPError {
	return httpError{err, statusCode}
}

var (
	ErrUnsupportedVersion               = NewHTTPError(errors.New("unsupported version"), http.StatusPreconditionFailed)
	ErrMaxSizeExceeded                  = NewHTTPError(errors.New("maximum size exceeded"), http.StatusRequestEntityTooLarge)
	ErrInvalidContentType               = NewHTTPError(errors.New("missing or invalid Content-Type header"), http.StatusBadRequest)
	ErrInvalidUploadLength              = NewHTTPError(errors.New("missing or invalid Upload-Length header"), http.StatusBadRequest)
	ErrInvalidOffset                    = NewHTTPError(errors.New("missing or invalid Upload-Offset header"), http.StatusBadRequest)
	ErrNotFound                         = NewHTTPError(errors.New("upload not found"), http.StatusNotFound)
	ErrFileLocked                       = NewHTTPError(errors.New("file currently locked"), 423) // Locked (WebDAV) (RFC 4918)
	ErrMismatchOffset                   = NewHTTPError(errors.New("mismatched offset"), http.StatusConflict)
	ErrSizeExceeded                     = NewHTTPError(errors.New("resource's size exceeded"), http.StatusRequestEntityTooLarge)
	ErrNotImplemented                   = NewHTTPError(errors.New("feature not implemented"), http.StatusNotImplemented)
	ErrUploadNotFinished                = NewHTTPError(errors.New("one of the partial uploads is not finished"), http.StatusBadRequest)
	ErrInvalidConcat                    = NewHTTPError(errors.New("invalid Upload-Concat header"), http.StatusBadRequest)
	ErrModifyFinal                      = NewHTTPError(errors.New("modifying a final upload is not allowed"), http.StatusForbidden)
	ErrUploadLengthAndUploadDeferLength = NewHTTPError(errors.New("provided both Upload-Length and Upload-Defer-Length"), http.StatusBadRequest)
	ErrInvalidUploadDeferLength         = NewHTTPError(errors.New("invalid Upload-Defer-Length header"), http.StatusBadRequest)
)

// UploadHandler exposes methods to handle requests as part of the tus protocol,
// such as PostFile, HeadFile, PatchFile and DelFile. In addition the GetFile method
// is provided which is, however, not part of the specification.
type UploadHandler struct {
	composer *tusd.StoreComposer
	http.Handler

	MaxSize int64
	*log.FieldLogger

	// CompleteUploads is used to send notifications whenever an upload is
	// completed by a user. The FileInfo will contain information about this
	// upload after it is completed. Sending to this channel will only
	// happen if the NotifyCompleteUploads field is set to true in the Config
	// structure. Notifications will also be sent for completions using the
	// Concatenation extension.
	NotifyCompleteUploads bool
	CompleteUploads       chan tusd.FileInfo
	// TerminatedUploads is used to send notifications whenever an upload is
	// terminated by a user. The FileInfo will contain information about this
	// upload gathered before the termination. Sending to this channel will only
	// happen if the NotifyTerminatedUploads field is set to true in the Config
	// structure.
	NotifyTerminatedUploads bool
	TerminatedUploads       chan tusd.FileInfo
	// UploadProgress is used to send notifications about the progress of the
	// currently running uploads. For each open PATCH request, every second
	// a FileInfo instance will be send over this channel with the Offset field
	// being set to the number of bytes which have been transfered to the server.
	// Please be aware that this number may be higher than the number of bytes
	// which have been stored by the data store! Sending to this channel will only
	// happen if the NotifyUploadProgress field is set to true in the Config
	// structure.
	NotifyUploadProgress bool
	UploadProgress       chan tusd.FileInfo
	// CreatedUploads is used to send notifications about the uploads having been
	// created. It triggers post creation and therefore has all the FileInfo incl.
	// the ID available already. It facilitates the post-create hook. Sending to
	// this channel will only happen if the NotifyCreatedUploads field is set to
	// true in the Config structure.
	NotifyCreatedUploads bool
	CreatedUploads       chan tusd.FileInfo
}

// NewUploadHandler creates a new handler without routing using the given
// configuration. It exposes the http handlers which need to be combined with
// a router (aka mux) of your choice. If you are looking for preconfigured
// handler see NewHandler.
func NewUploadHandler(config tusd.Config) (*UploadHandler, error) {
	l := log.New(nil)
	l.SetStdLogger(config.Logger)
	handler := &UploadHandler{
		composer:                config.StoreComposer,
		CompleteUploads:         make(chan tusd.FileInfo),
		TerminatedUploads:       make(chan tusd.FileInfo),
		UploadProgress:          make(chan tusd.FileInfo),
		CreatedUploads:          make(chan tusd.FileInfo),
		FieldLogger:             l,
		NotifyCompleteUploads:   config.NotifyCompleteUploads,
		NotifyTerminatedUploads: config.NotifyTerminatedUploads,
		NotifyUploadProgress:    config.NotifyUploadProgress,
		NotifyCreatedUploads:    config.NotifyCreatedUploads,
		MaxSize:                 config.MaxSize,
	}

	handler.Handler = handler.newHandler()
	return handler, nil
}
func (handler *UploadHandler) newHandler() http.Handler {
	m := pat.New()
	h := handler.Middleware(m)

	m.Post("", http.HandlerFunc(handler.PostFile))
	m.Head(":id", http.HandlerFunc(handler.HeadFile))
	m.Patch(":id", http.HandlerFunc(handler.PatchFile))
	m.Del(":id", http.HandlerFunc(handler.DelFile))
	m.Get(":id", http.HandlerFunc(handler.GetFile))
	return h
}

// Middleware checks various aspects of the request and ensures that it
// conforms with the spec. Also handles method overriding for clients which
// cannot make PATCH AND DELETE requests. If you are using the tusd handlers
// directly you will need to wrap at least the POST and PATCH endpoints in
// this middleware.
func (handler *UploadHandler) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow overriding the HTTP method. The reason for this is
		// that some libraries/environments to not support PATCH and
		// DELETE requests, e.g. Flash in a browser and parts of Java
		if newMethod := r.Header.Get("X-HTTP-Method-Override"); newMethod != "" {
			r.Method = newMethod
		}

		handler.GetLogger().WithField("method", r.Method).WithField("path", r.URL.Path).Info("RequestIncoming")

		header := w.Header()

		if origin := r.Header.Get("Origin"); origin != "" {
			header.Set("Access-Control-Allow-Origin", origin)

			if r.Method == "OPTIONS" {
				// Preflight request
				header.Add("Access-Control-Allow-Methods", "POST, GET, HEAD, PATCH, DELETE, OPTIONS")
				header.Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Upload-Length, Upload-Offset, Tus-Resumable, Upload-Metadata, Upload-Defer-Length, Upload-Concat")
				header.Set("Access-Control-Max-Age", "86400")
			} else {
				// Actual request
				header.Add("Access-Control-Expose-Headers", "Upload-Offset, Location, Upload-Length, Tus-Version, Tus-Resumable, Tus-Max-Size, Tus-Extension, Upload-Metadata, Upload-Defer-Length, Upload-Concat")
			}
		}

		// Add nosniff to all responses https://golang.org/src/net/http/server.go#L1429
		header.Set("X-Content-Type-Options", "nosniff")

		// Set appropriated headers in case of OPTIONS method allowing protocol
		// discovery and end with an 204 No Content
		if r.Method == "OPTIONS" {
			if handler.MaxSize > 0 {
				header.Set("Tus-Max-Size", strconv.FormatInt(handler.MaxSize, 10))
			}

			// Although the 204 No Content status code is a better fit in this case,
			// since we do not have a response body included, we cannot use it here
			// as some browsers only accept 200 OK as successful response to a
			// preflight request. If we send them the 204 No Content the response
			// will be ignored or interpreted as a rejection.
			// For example, the Presto engine, which is used in older versions of
			// Opera, Opera Mobile and Opera Mini, handles CORS this way.
			handler.sendResp(w, r, http.StatusOK)
			return
		}

		// Proceed with routing the request
		h.ServeHTTP(w, r)
	})
}

// PostFile creates a new file upload using the datastore after validating the
// length and parsing the metadata.
func (handler *UploadHandler) PostFile(w http.ResponseWriter, r *http.Request) {

	info := tusd.FileInfo{
		SizeIsDeferred: true,
		IsFinal:        true,
	}

	id, err := handler.composer.Core.NewUpload(info)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	info.ID = id

	// Add the Location header directly after creating the new resource to even
	// include it in cases of failure when an error is returned
	url := handler.absFileURL(r, id)
	w.Header().Set("Location", url)

	if handler.NotifyCreatedUploads {
		handler.CreatedUploads <- info
	}

	if err := handler.composer.Concater.ConcatUploads(id, []string{}); err != nil {
		handler.sendError(w, r, err)
		return
	}

	if handler.NotifyCompleteUploads {
		handler.CompleteUploads <- info
	}

	if handler.composer.UsesLocker {
		locker := handler.composer.Locker
		if err := locker.LockUpload(id); err != nil {
			handler.sendError(w, r, err)
			return
		}

		defer locker.UnlockUpload(id)
	}

	if err := handler.writeChunk(id, info, w, r); err != nil {
		handler.sendError(w, r, err)
		return
	}

	handler.sendResp(w, r, http.StatusCreated)
}

// HeadFile returns the length and offset for the HEAD request
func (handler *UploadHandler) HeadFile(w http.ResponseWriter, r *http.Request) {

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	if handler.composer.UsesLocker {
		locker := handler.composer.Locker
		if err := locker.LockUpload(id); err != nil {
			handler.sendError(w, r, err)
			return
		}

		defer locker.UnlockUpload(id)
	}

	info, err := handler.composer.Core.GetInfo(id)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	// Add Upload-Concat header if possible
	if info.IsPartial {
		w.Header().Set("Upload-Concat", "partial")
	}

	if info.IsFinal {
		v := "final;"
		for _, uploadID := range info.PartialUploads {
			v += handler.absFileURL(r, uploadID) + " "
		}
		// Remove trailing space
		v = v[:len(v)-1]

		w.Header().Set("Upload-Concat", v)
	}

	if len(info.MetaData) != 0 {
		w.Header().Set("Upload-Metadata", SerializeMetadataHeader(info.MetaData))
	}

	if info.SizeIsDeferred {
		w.Header().Set("Upload-Defer-Length", UploadLengthDeferred)
	} else {
		w.Header().Set("Upload-Length", strconv.FormatInt(info.Size, 10))
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Upload-Offset", strconv.FormatInt(info.Offset, 10))
	handler.sendResp(w, r, http.StatusOK)
}

// PatchFile adds a chunk to an upload. This operation is only allowed
// if enough space in the upload is left.
func (handler *UploadHandler) PatchFile(w http.ResponseWriter, r *http.Request) {

	// Check for presence of application/offset+octet-stream
	if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
		handler.sendError(w, r, ErrInvalidContentType)
		return
	}

	// Check for presence of a valid Upload-Offset Header
	offset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil || offset < 0 {
		handler.sendError(w, r, ErrInvalidOffset)
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	if handler.composer.UsesLocker {
		locker := handler.composer.Locker
		if err := locker.LockUpload(id); err != nil {
			handler.sendError(w, r, err)
			return
		}

		defer locker.UnlockUpload(id)
	}

	info, err := handler.composer.Core.GetInfo(id)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	// Modifying a final upload is not allowed
	if info.IsFinal {
		handler.sendError(w, r, ErrModifyFinal)
		return
	}

	if offset != info.Offset {
		handler.sendError(w, r, ErrMismatchOffset)
		return
	}

	// Do not proxy the call to the data store if the upload is already completed
	if !info.SizeIsDeferred && info.Offset == info.Size {
		w.Header().Set("Upload-Offset", strconv.FormatInt(offset, 10))
		handler.sendResp(w, r, http.StatusNoContent)
		return
	}

	if r.Header.Get("Upload-Length") != "" {
		if !handler.composer.UsesLengthDeferrer {
			handler.sendError(w, r, ErrNotImplemented)
			return
		}
		if !info.SizeIsDeferred {
			handler.sendError(w, r, ErrInvalidUploadLength)
			return
		}
		uploadLength, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
		if err != nil || uploadLength < 0 || uploadLength < info.Offset || (handler.MaxSize > 0 && uploadLength > handler.MaxSize) {
			handler.sendError(w, r, ErrInvalidUploadLength)
			return
		}
		if err := handler.composer.LengthDeferrer.DeclareLength(id, uploadLength); err != nil {
			handler.sendError(w, r, err)
			return
		}

		info.Size = uploadLength
		info.SizeIsDeferred = false
	}

	if err := handler.writeChunk(id, info, w, r); err != nil {
		handler.sendError(w, r, err)
		return
	}

	handler.sendResp(w, r, http.StatusNoContent)
}

// writeChunk reads the body from the requests r and appends it to the upload
// with the corresponding id. Afterwards, it will set the necessary response
// headers but will not send the response.
func (handler *UploadHandler) writeChunk(id string, info tusd.FileInfo, w http.ResponseWriter, r *http.Request) error {
	// Get Content-Length if possible
	length := r.ContentLength
	offset := info.Offset

	// Test if this upload fits into the file's size
	if !info.SizeIsDeferred && offset+length > info.Size {
		return ErrSizeExceeded
	}

	maxSize := info.Size - offset
	// If the upload's length is deferred and the PATCH request does not contain the Content-Length
	// header (which is allowed if 'Transfer-Encoding: chunked' is used), we still need to set limits for
	// the body size.
	if info.SizeIsDeferred {
		if handler.MaxSize > 0 {
			// Ensure that the upload does not exceed the maximum upload size
			maxSize = handler.MaxSize - offset
		} else {
			// If no upload limit is given, we allow arbitrary sizes
			maxSize = math.MaxInt64
		}
	}
	if length > 0 {
		maxSize = length
	}

	var bytesWritten int64
	// Prevent a nil pointer dereference when accessing the body which may not be
	// available in the case of a malicious request.
	if r.Body != nil {
		var uploadFile io.ReadCloser
		uploadFile, _, err := r.FormFile("file")
		if err != nil {
			if err != http.ErrMissingFile && err != http.ErrNotMultipart {
				return err
			}
			uploadFile = r.Body
		}
		defer uploadFile.Close()

		// Limit the data read from the request's body to the allowed maximum
		reader := io.LimitReader(uploadFile, maxSize)

		if handler.NotifyUploadProgress {
			var stop chan<- struct{}
			reader, stop = handler.sendProgressMessages(info, reader)
			defer close(stop)
		}

		bytesWritten, err = handler.composer.Core.WriteChunk(id, offset, reader)
		if err != nil {
			return err
		}
	}

	// Send new offset to client
	newOffset := offset + bytesWritten
	w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))
	info.Offset = newOffset

	return handler.finishUploadIfComplete(info)
}

// finishUploadIfComplete checks whether an upload is completed (i.e. upload offset
// matches upload size) and if so, it will call the data store's FinishUpload
// function and send the necessary message on the CompleteUpload channel.
func (handler *UploadHandler) finishUploadIfComplete(info tusd.FileInfo) error {
	// If the upload is completed, ...
	if !info.SizeIsDeferred && info.Offset == info.Size {
		// ... allow custom mechanism to finish and cleanup the upload
		if handler.composer.UsesFinisher {
			if err := handler.composer.Finisher.FinishUpload(info.ID); err != nil {
				return err
			}
		}

		// ... send the info out to the channel
		if handler.NotifyCompleteUploads {
			handler.CompleteUploads <- info
		}
	}

	return nil
}

// GetFile handles requests to download a file using a GET request. This is not
// part of the specification.
func (handler *UploadHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	if !handler.composer.UsesGetReader {
		handler.sendError(w, r, ErrNotImplemented)
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	if handler.composer.UsesLocker {
		locker := handler.composer.Locker
		if err := locker.LockUpload(id); err != nil {
			handler.sendError(w, r, err)
			return
		}

		defer locker.UnlockUpload(id)
	}

	info, err := handler.composer.Core.GetInfo(id)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	// Set headers before sending responses
	w.Header().Set("Content-Length", strconv.FormatInt(info.Offset, 10))

	contentType, contentDisposition := filterContentType(info)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", contentDisposition)

	// If no data has been uploaded yet, respond with an empty "204 No Content" status.
	if info.Offset == 0 {
		handler.sendResp(w, r, http.StatusNoContent)
		return
	}

	src, err := handler.composer.GetReader.GetReader(id)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	handler.sendResp(w, r, http.StatusOK)
	io.Copy(w, src)

	// Try to close the reader if the io.Closer interface is implemented
	if closer, ok := src.(io.Closer); ok {
		closer.Close()
	}
}

// mimeInlineBrowserWhitelist is a map containing MIME types which should be
// allowed to be rendered by browser inline, instead of being forced to be
// downloadd. For example, HTML or SVG files are not allowed, since they may
// contain malicious JavaScript. In a similiar fashion PDF is not on this list
// as their parsers commonly contain vulnerabilities which can be exploited.
// The values of this map does not convei any meaning and are therefore just
// empty structs.
var mimeInlineBrowserWhitelist = map[string]struct{}{
	"text/plain": struct{}{},

	"image/png":  struct{}{},
	"image/jpeg": struct{}{},
	"image/gif":  struct{}{},
	"image/bmp":  struct{}{},
	"image/webp": struct{}{},

	"audio/wave":      struct{}{},
	"audio/wav":       struct{}{},
	"audio/x-wav":     struct{}{},
	"audio/x-pn-wav":  struct{}{},
	"audio/webm":      struct{}{},
	"video/webm":      struct{}{},
	"audio/ogg":       struct{}{},
	"video/ogg ":      struct{}{},
	"application/ogg": struct{}{},
}

// filterContentType returns the values for the Content-Type and
// Content-Disposition headers for a given upload. These values should be used
// in responses for GET requests to ensure that only non-malicious file types
// are shown directly in the browser. It will extract the file name and type
// from the "fileame" and "filetype".
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition
func filterContentType(info tusd.FileInfo) (contentType string, contentDisposition string) {
	filetype := info.MetaData["filetype"]

	if reMimeType.MatchString(filetype) {
		// If the filetype from metadata is well formed, we forward use this
		// for the Content-Type header. However, only whitelisted mime types
		// will be allowed to be shown inline in the browser
		contentType = filetype
		if _, isWhitelisted := mimeInlineBrowserWhitelist[filetype]; isWhitelisted {
			contentDisposition = "inline"
		} else {
			contentDisposition = "attachment"
		}
	} else {
		// If the filetype from the metadata is not well formed, we use a
		// default type and force the browser to download the content.
		contentType = "application/octet-stream"
		contentDisposition = "attachment"
	}

	// Add a filename to Content-Disposition if one is available in the metadata
	if filename, ok := info.MetaData["filename"]; ok {
		contentDisposition += ";filename=" + strconv.Quote(filename)
	}

	return contentType, contentDisposition
}

// DelFile terminates an upload permanently.
func (handler *UploadHandler) DelFile(w http.ResponseWriter, r *http.Request) {
	// Abort the request handling if the required interface is not implemented
	if !handler.composer.UsesTerminater {
		handler.sendError(w, r, ErrNotImplemented)
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	if handler.composer.UsesLocker {
		locker := handler.composer.Locker
		if err := locker.LockUpload(id); err != nil {
			handler.sendError(w, r, err)
			return
		}

		defer locker.UnlockUpload(id)
	}

	var info tusd.FileInfo
	if handler.NotifyTerminatedUploads {
		info, err = handler.composer.Core.GetInfo(id)
		if err != nil {
			handler.sendError(w, r, err)
			return
		}
	}

	err = handler.composer.Terminater.Terminate(id)
	if err != nil {
		handler.sendError(w, r, err)
		return
	}

	handler.sendResp(w, r, http.StatusNoContent)

	if handler.NotifyTerminatedUploads {
		handler.TerminatedUploads <- info
	}

}

// Send the error in the response body. The status code will be looked up in
// ErrStatusCodes. If none is found 500 Internal Error will be used.
func (handler *UploadHandler) sendError(w http.ResponseWriter, r *http.Request, err error) {
	// Interpret os.ErrNotExist as 404 Not Found
	if os.IsNotExist(err) {
		err = ErrNotFound
	}

	// Errors for read timeouts contain too much information which is not
	// necessary for us and makes grouping for the metrics harder. The error
	// message looks like: read tcp 127.0.0.1:1080->127.0.0.1:53673: i/o timeout
	// Therefore, we use a common error message for all of them.
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		err = errors.New("read tcp: i/o timeout")
	}

	statusErr, ok := err.(HTTPError)
	if !ok {
		statusErr = NewHTTPError(err, http.StatusInternalServerError)
	}

	reason := append(statusErr.Body(), '\n')
	if r.Method == "HEAD" {
		reason = nil
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(reason)))
	w.WriteHeader(statusErr.StatusCode())
	w.Write(reason)
}

// sendResp writes the header to w with the specified status code.
func (handler *UploadHandler) sendResp(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
}

// Make an absolute URLs to the given upload id. If the base path is absolute
// it will be prepended else the host and protocol from the request is used.
func (handler *UploadHandler) absFileURL(r *http.Request, id string) string {

	scheme, host := getSchemeAndHost(r, true)
	url, _ := url.Parse(path.Join(r.URL.Path, id))
	url = r.URL.ResolveReference(url)
	url.Host = host
	url.Scheme = scheme
	return url.String()
}

type progressWriter struct {
	Offset int64
}

func (w *progressWriter) Write(b []byte) (int, error) {
	atomic.AddInt64(&w.Offset, int64(len(b)))
	return len(b), nil
}

// sendProgressMessage will send a notification over the UploadProgress channel
// every second, indicating how much data has been transfered to the server.
// It will stop sending these instances once the returned channel has been
// closed. The returned reader should be used to read the request body.
func (handler *UploadHandler) sendProgressMessages(info tusd.FileInfo, reader io.Reader) (io.Reader, chan<- struct{}) {
	previousOffset := int64(0)
	progress := &progressWriter{
		Offset: info.Offset,
	}
	stop := make(chan struct{}, 1)
	reader = io.TeeReader(reader, progress)

	go func() {
		for {
			select {
			case <-stop:
				info.Offset = atomic.LoadInt64(&progress.Offset)
				if info.Offset != previousOffset {
					handler.UploadProgress <- info
					previousOffset = info.Offset
				}
				return
			case <-time.After(1 * time.Second):
				info.Offset = atomic.LoadInt64(&progress.Offset)
				if info.Offset != previousOffset {
					handler.UploadProgress <- info
					previousOffset = info.Offset
				}
			}
		}
	}()

	return reader, stop
}

// getHostAndProtocol extracts the host and used protocol (either HTTP or HTTPS)
// from the given request. If `allowForwarded` is set, the X-Forwarded-Host,
// X-Forwarded-Proto and Forwarded headers will also be checked to
// support proxies.
func getHostAndProtocol(r *http.Request, allowForwarded bool) (host, proto string) {
	if r.TLS != nil {
		proto = "https"
	} else {
		proto = "http"
	}

	host = r.Host

	if !allowForwarded {
		return
	}

	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}

	if h := r.Header.Get("X-Forwarded-Proto"); h == "http" || h == "https" {
		proto = h
	}

	if h := r.Header.Get("Forwarded"); h != "" {
		if r := reForwardedHost.FindStringSubmatch(h); len(r) == 2 {
			host = r[1]
		}

		if r := reForwardedProto.FindStringSubmatch(h); len(r) == 2 {
			proto = r[1]
		}
	}

	return
}

// The get sum of all sizes for a list of upload ids while checking whether
// all of these uploads are finished yet. This is used to calculate the size
// of a final resource.
func (handler *UploadHandler) sizeOfUploads(ids []string) (size int64, err error) {
	for _, id := range ids {
		info, err := handler.composer.Core.GetInfo(id)
		if err != nil {
			return size, err
		}

		if info.SizeIsDeferred || info.Offset != info.Size {
			err = ErrUploadNotFinished
			return size, err
		}

		size += info.Size
	}

	return
}

// Verify that the Upload-Length and Upload-Defer-Length headers are acceptable for creating a
// new upload
func (handler *UploadHandler) validateNewUploadLengthHeaders(uploadLengthHeader string, uploadDeferLengthHeader string) (uploadLength int64, uploadLengthDeferred bool, err error) {
	haveBothLengthHeaders := uploadLengthHeader != "" && uploadDeferLengthHeader != ""
	haveInvalidDeferHeader := uploadDeferLengthHeader != "" && uploadDeferLengthHeader != UploadLengthDeferred
	lengthIsDeferred := uploadDeferLengthHeader == UploadLengthDeferred

	if lengthIsDeferred && !handler.composer.UsesLengthDeferrer {
		err = ErrNotImplemented
	} else if haveBothLengthHeaders {
		err = ErrUploadLengthAndUploadDeferLength
	} else if haveInvalidDeferHeader {
		err = ErrInvalidUploadDeferLength
	} else if lengthIsDeferred {
		uploadLengthDeferred = true
	} else {
		uploadLength, err = strconv.ParseInt(uploadLengthHeader, 10, 64)
		if err != nil || uploadLength < 0 {
			err = ErrInvalidUploadLength
		}
	}

	return
}

// ParseMetadataHeader parses the Upload-Metadata header as defined in the
// File Creation extension.
// e.g. Upload-Metadata: name bHVucmpzLnBuZw==,type aW1hZ2UvcG5n
func ParseMetadataHeader(header string) map[string]string {
	meta := make(map[string]string)

	for _, element := range strings.Split(header, ",") {
		element := strings.TrimSpace(element)

		parts := strings.Split(element, " ")

		// Do not continue with this element if no key and value or presented
		if len(parts) != 2 {
			continue
		}

		// Ignore corrent element if the value is no valid base64
		key := parts[0]
		value, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			continue
		}

		meta[key] = string(value)
	}

	return meta
}

// SerializeMetadataHeader serializes a map of strings into the Upload-Metadata
// header format used in the response for HEAD requests.
// e.g. Upload-Metadata: name bHVucmpzLnBuZw==,type aW1hZ2UvcG5n
func SerializeMetadataHeader(meta map[string]string) string {
	header := ""
	for key, value := range meta {
		valueBase64 := base64.StdEncoding.EncodeToString([]byte(value))
		header += key + " " + valueBase64 + ","
	}

	// Remove trailing comma
	if len(header) > 0 {
		header = header[:len(header)-1]
	}

	return header
}

// Parse the Upload-Concat header, e.g.
// Upload-Concat: partial
// Upload-Concat: final;http://tus.io/files/a /files/b/
func parseConcat(header string) (isPartial bool, isFinal bool, partialUploads []string, err error) {
	if len(header) == 0 {
		return
	}

	if header == "partial" {
		isPartial = true
		return
	}

	l := len("final;")
	if strings.HasPrefix(header, "final;") && len(header) > l {
		isFinal = true

		list := strings.Split(header[l:], " ")
		for _, value := range list {
			value := strings.TrimSpace(value)
			if value == "" {
				continue
			}

			id, extractErr := extractIDFromPath(value)
			if extractErr != nil {
				err = extractErr
				return
			}

			partialUploads = append(partialUploads, id)
		}
	}

	// If no valid partial upload ids are extracted this is not a final upload.
	if len(partialUploads) == 0 {
		isFinal = false
		err = ErrInvalidConcat
	}

	return
}

// extractIDFromPath pulls the last segment from the url provided
func extractIDFromPath(url string) (string, error) {
	result := reExtractFileID.FindStringSubmatch(url)
	if len(result) != 2 {
		return "", ErrNotFound
	}
	return result[1], nil
}

func i64toa(num int64) string {
	return strconv.FormatInt(num, 10)
}

// getHostAndProtocol extracts the host and used protocol (either HTTP or HTTPS)
// from the given request. If `allowForwarded` is set, the X-Forwarded-Host,
// X-Forwarded-Proto and Forwarded headers will also be checked to
// support proxies.
func getSchemeAndHost(r *http.Request, allowForwarded bool) (scheme, host string) {
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	host = r.Host

	if !allowForwarded {
		return
	}

	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}

	if h := r.Header.Get("X-Forwarded-Proto"); h == "http" || h == "https" {
		scheme = h
	}

	if h := r.Header.Get("Forwarded"); h != "" {
		if r := reForwardedHost.FindStringSubmatch(h); len(r) == 2 {
			host = r[1]
		}

		if r := reForwardedProto.FindStringSubmatch(h); len(r) == 2 {
			scheme = r[1]
		}
	}

	return
}
