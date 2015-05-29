package storage

import (
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
	"github.com/docker/libtrust"
)

type manifestStore struct {
	repository    *repository
	revisionStore *revisionStore
	tagStore      *tagStore
	ctx           context.Context
}

var _ distribution.ManifestService = &manifestStore{}

func (ms *manifestStore) Exists(dgst digest.Digest) (bool, error) {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).Exists")

	_, err := ms.revisionStore.blobStore.Stat(ms.ctx, dgst)
	if err != nil {
		if err == distribution.ErrBlobUnknown {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (ms *manifestStore) Get(dgst digest.Digest) (*manifest.SignedManifest, error) {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).Get")
	return ms.revisionStore.get(ms.ctx, dgst)
}

func (ms *manifestStore) Put(manifest *manifest.SignedManifest) error {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).Put")

	// Store the revision of the manifest
	revision, err := ms.revisionStore.put(ms.ctx, manifest)
	if err != nil {
		return err
	}

	// Now, tag the manifest
	return ms.tagStore.tag(manifest.Tag, revision.Digest)
}

// Delete removes the revision of the specified manfiest.
func (ms *manifestStore) Delete(dgst digest.Digest) error {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).Delete - unsupported")
	return fmt.Errorf("deletion of manifests not supported")
}

func (ms *manifestStore) Tags() ([]string, error) {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).Tags")
	return ms.tagStore.tags()
}

func (ms *manifestStore) ExistsByTag(tag string) (bool, error) {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).ExistsByTag")
	return ms.tagStore.exists(tag)
}

func (ms *manifestStore) GetByTag(tag string) (*manifest.SignedManifest, error) {
	context.GetLogger(ms.ctx).Debug("(*manifestStore).GetByTag")
	dgst, err := ms.tagStore.resolve(tag)
	if err != nil {
		return nil, err
	}

	return ms.revisionStore.get(ms.ctx, dgst)
}

// VerifyLocalManifest ensures that the manifest content is valid from the
// perspective of the registry. It ensures that the signature is valid for the
// enclosed payload. As a policy, the registry only tries to store valid
// content, leaving trust policies of that content up to consumers.
func (ms *manifestStore) Verify(ctx context.Context, mnfst *manifest.SignedManifest) error {
	var errs distribution.ErrManifestVerification
	if mnfst.Name != ms.repository.Name() {
		errs = append(errs, fmt.Errorf("repository name does not match manifest name"))
	}

	if _, err := manifest.Verify(mnfst); err != nil {
		switch err {
		case libtrust.ErrMissingSignatureKey, libtrust.ErrInvalidJSONContent, libtrust.ErrMissingSignatureKey:
			errs = append(errs, distribution.ErrManifestUnverified{})
		default:
			if err.Error() == "invalid signature" { // TODO(stevvooe): This should be exported by libtrust
				errs = append(errs, distribution.ErrManifestUnverified{})
			} else {
				errs = append(errs, err)
			}
		}
	}

	for _, fsLayer := range mnfst.FSLayers {
		_, err := ms.repository.Blobs(ctx).Stat(ctx, fsLayer.BlobSum)
		if err != nil {
			if err != distribution.ErrBlobUnknown {
				errs = append(errs, err)
			}

			// On error here, we always append unknown blob errors.
			errs = append(errs, distribution.ErrManifestBlobUnknown{Digest: fsLayer.BlobSum})
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}
