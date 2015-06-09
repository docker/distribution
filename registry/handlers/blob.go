package handlers

import (
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/storage"
	"github.com/gorilla/handlers"
)

// blobDispatcher uses the request context to build a blobHandler.
func blobDispatcher(ctx *Context, r *http.Request) http.Handler {
	dgst, err := getDigest(ctx)
	if err != nil {

		if err == errDigestNotAvailable {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				ctx.Errors.Push(v2.ErrorCodeDigestInvalid, err)
			})
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Errors.Push(v2.ErrorCodeDigestInvalid, err)
		})
	}

	blobHandler := &blobHandler{
		Context: ctx,
		Digest:  dgst,
	}

	return handlers.MethodHandler{
		"GET":    http.HandlerFunc(blobHandler.GetBlob),
		"HEAD":   http.HandlerFunc(blobHandler.GetBlob),
		"DELETE": http.HandlerFunc(blobHandler.DeleteBlob),
	}
}

// blobHandler serves http blob requests.
type blobHandler struct {
	*Context

	Digest digest.Digest
}

// GetBlob fetches the binary data from backend storage returns it in the
// response.
func (bh *blobHandler) GetBlob(w http.ResponseWriter, r *http.Request) {
	context.GetLogger(bh).Debug("GetBlob")
	blobs := bh.Repository.Blobs(bh)
	desc, err := blobs.Stat(bh, bh.Digest)
	if err != nil {
		if err == distribution.ErrBlobUnknown {
			w.WriteHeader(http.StatusNotFound)
			bh.Errors.Push(v2.ErrorCodeBlobUnknown, bh.Digest)
		} else {
			bh.Errors.Push(v2.ErrorCodeUnknown, err)
		}
		return
	}

	if err := blobs.ServeBlob(bh, w, r, desc.Digest); err != nil {
		context.GetLogger(bh).Debugf("unexpected error getting blob HTTP handler: %v", err)
		bh.Errors.Push(v2.ErrorCodeUnknown, err)
		return
	}
}

// DeleteBlob deletes a layer blob
func (bh *blobHandler) DeleteBlob(w http.ResponseWriter, r *http.Request) {
	context.GetLogger(bh).Debug("DeleteBlob")

	blobs := bh.Repository.Blobs(bh)
	err := blobs.Delete(bh, bh.Digest)
	if err != nil {
		switch err {
		case distribution.ErrBlobUnknown:
			w.WriteHeader(http.StatusNotFound)
			bh.Errors.Push(v2.ErrorCodeBlobUnknown, bh.Digest)
		case storage.ErrDeleteDisabled:
			w.WriteHeader(http.StatusMethodNotAllowed)
			bh.Errors.Push(v2.ErrorCodeDisabled, bh.Digest)
		default:
			bh.Errors.Push(v2.ErrorCodeUnknown, err)
		}
		return
	}

	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusAccepted)
}
