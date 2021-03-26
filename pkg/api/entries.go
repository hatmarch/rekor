/*
Copyright © 2020 Bob Callaway <bcallawa@redhat.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/trillian"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/go-openapi/swag"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"

	"github.com/sigstore/rekor/pkg/log"
	"github.com/sigstore/rekor/pkg/types"

	"github.com/sigstore/rekor/pkg/generated/models"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	tclient "github.com/google/trillian/client"
	tcrypto "github.com/google/trillian/crypto"
	rfc6962 "github.com/google/trillian/merkle/rfc6962/hasher"
	"github.com/sigstore/rekor/pkg/generated/restapi/operations/entries"
)

func GetLogEntryByIndexHandler(params entries.GetLogEntryByIndexParams) middleware.Responder {
	tc := NewTrillianClient(params.HTTPRequest.Context())

	resp := tc.getLeafByIndex(params.LogIndex)
	switch resp.status {
	case codes.OK:
	case codes.NotFound, codes.OutOfRange:
		return handleRekorAPIError(params, http.StatusNotFound, fmt.Errorf("grpc error: %w", resp.err), "")
	default:
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc err: %w", resp.err), trillianCommunicationError)
	}

	leaves := resp.getLeafByRangeResult.GetLeaves()
	if len(leaves) > 1 {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("len(leaves): %v", len(leaves)), trillianUnexpectedResult)
	} else if len(leaves) == 0 {
		return handleRekorAPIError(params, http.StatusNotFound, errors.New("grpc returned 0 leaves with success code"), "")
	}
	leaf := leaves[0]

	logEntry := models.LogEntry{
		hex.EncodeToString(leaf.MerkleLeafHash): models.LogEntryAnon{
			LogIndex:       &leaf.LeafIndex,
			Body:           leaf.LeafValue,
			IntegratedTime: leaf.IntegrateTimestamp.AsTime().Unix(),
		},
	}
	return entries.NewGetLogEntryByIndexOK().WithPayload(logEntry)
}

func CreateLogEntryHandler(params entries.CreateLogEntryParams) middleware.Responder {
	httpReq := params.HTTPRequest
	entry, err := types.NewEntry(params.ProposedEntry)
	if err != nil {
		return handleRekorAPIError(params, http.StatusBadRequest, err, err.Error())
	}

	leaf, err := entry.Canonicalize(httpReq.Context())
	if err != nil {
		return handleRekorAPIError(params, http.StatusInternalServerError, err, failedToGenerateCanonicalEntry)
	}

	tc := NewTrillianClient(httpReq.Context())

	resp := tc.addLeaf(leaf)
	//this represents overall GRPC response state (not the results of insertion into the log)
	if resp.status != codes.OK {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianUnexpectedResult)
	}

	//this represents the results of inserting the proposed leaf into the log; status is nil in success path
	insertionStatus := resp.getAddResult.QueuedLeaf.Status
	if insertionStatus != nil {
		switch insertionStatus.Code {
		case int32(code.Code_OK):
		case int32(code.Code_ALREADY_EXISTS), int32(code.Code_FAILED_PRECONDITION):
			existingUUID := hex.EncodeToString(rfc6962.DefaultHasher.HashLeaf(leaf))
			return handleRekorAPIError(params, http.StatusConflict, fmt.Errorf("grpc error: %v", insertionStatus.String()), fmt.Sprintf(entryAlreadyExists, existingUUID), "entryURL", getEntryURL(*httpReq.URL, existingUUID))
		default:
			return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %v", insertionStatus.String()), trillianUnexpectedResult)
		}
	}

	// We made it this far, that means the entry was successfully added.
	metricNewEntries.Inc()

	queuedLeaf := resp.getAddResult.QueuedLeaf.Leaf
	uuid := hex.EncodeToString(queuedLeaf.GetMerkleLeafHash())

	logEntry := models.LogEntry{
		uuid: models.LogEntryAnon{
			LogIndex: swag.Int64(queuedLeaf.LeafIndex),
			Body:     queuedLeaf.GetLeafValue(),
		},
	}

	if viper.GetBool("enable_retrieve_api") {
		go func() {
			for _, key := range entry.IndexKeys() {
				if err := addToIndex(context.Background(), key, uuid); err != nil {
					log.RequestIDLogger(params.HTTPRequest).Error(err)
				}
			}
		}()
	}

	return entries.NewCreateLogEntryCreated().WithPayload(logEntry).WithLocation(getEntryURL(*httpReq.URL, uuid)).WithETag(uuid)
}

func getEntryURL(locationURL url.URL, uuid string) strfmt.URI {
	// remove API key from output
	query := locationURL.Query()
	query.Del("apiKey")
	locationURL.RawQuery = query.Encode()
	locationURL.Path = fmt.Sprintf("%v/%v", locationURL.Path, uuid)
	return strfmt.URI(locationURL.String())

}

func GetLogEntryByUUIDHandler(params entries.GetLogEntryByUUIDParams) middleware.Responder {
	hashValue, _ := hex.DecodeString(params.EntryUUID)
	hashes := [][]byte{hashValue}

	tc := NewTrillianClient(params.HTTPRequest.Context())

	resp := tc.getLeafByHash(hashes) // TODO: if this API is deprecated, we need to ask for inclusion proof and then use index in proof result to get leaf
	switch resp.status {
	case codes.OK:
	case codes.NotFound:
		return handleRekorAPIError(params, http.StatusNotFound, fmt.Errorf("grpc error: %w", resp.err), "")
	default:
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianUnexpectedResult)
	}

	leaves := resp.getLeafResult.GetLeaves()
	if len(leaves) > 1 {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("len(leaves): %v", len(leaves)), trillianUnexpectedResult)
	} else if len(leaves) == 0 {
		return handleRekorAPIError(params, http.StatusNotFound, errors.New("grpc returned 0 leaves with success code"), "")
	}
	leaf := leaves[0]

	uuid := hex.EncodeToString(leaf.GetMerkleLeafHash())

	logEntry := models.LogEntry{
		uuid: models.LogEntryAnon{
			LogIndex:       swag.Int64(leaf.GetLeafIndex()),
			Body:           leaf.LeafValue,
			IntegratedTime: leaf.IntegrateTimestamp.AsTime().Unix(),
		},
	}
	return entries.NewGetLogEntryByUUIDOK().WithPayload(logEntry)
}

func GetLogEntryProofHandler(params entries.GetLogEntryProofParams) middleware.Responder {
	hashValue, _ := hex.DecodeString(params.EntryUUID)
	tc := NewTrillianClient(params.HTTPRequest.Context())

	resp := tc.getProofByHash(hashValue)
	switch resp.status {
	case codes.OK:
	case codes.NotFound:
		return handleRekorAPIError(params, http.StatusNotFound, fmt.Errorf("grpc error: %w", resp.err), "")
	default:
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianUnexpectedResult)
	}
	result := resp.getProofResult

	// validate result is signed with the key we're aware of
	pub, err := x509.ParsePKIXPublicKey(tc.pubkey.Der)
	if err != nil {
		return handleRekorAPIError(params, http.StatusInternalServerError, err, "")
	}
	verifier := tclient.NewLogVerifier(rfc6962.DefaultHasher, pub, crypto.SHA256)
	root, err := tcrypto.VerifySignedLogRoot(verifier.PubKey, verifier.SigHash, result.SignedLogRoot)
	if err != nil {
		return handleRekorAPIError(params, http.StatusInternalServerError, err, trillianUnexpectedResult)
	}

	if len(result.Proof) != 1 {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("len(result.Proof) = %v", len(result.Proof)), trillianUnexpectedResult)
	}
	proof := result.Proof[0]

	hashes := []string{}
	for _, hash := range proof.Hashes {
		hashes = append(hashes, hex.EncodeToString(hash))
	}

	inclusionProof := models.InclusionProof{
		TreeSize: swag.Int64(int64(root.TreeSize)),
		RootHash: swag.String(hex.EncodeToString(root.RootHash)),
		LogIndex: swag.Int64(proof.GetLeafIndex()),
		Hashes:   hashes,
	}
	return entries.NewGetLogEntryProofOK().WithPayload(&inclusionProof)
}

func SearchLogQueryHandler(params entries.SearchLogQueryParams) middleware.Responder {
	httpReqCtx := params.HTTPRequest.Context()
	resultPayload := []models.LogEntry{}
	tc := NewTrillianClient(httpReqCtx)

	if len(params.Entry.EntryUUIDs) > 0 || len(params.Entry.Entries()) > 0 {
		g, _ := errgroup.WithContext(httpReqCtx)

		searchHashes := make([][]byte, len(params.Entry.EntryUUIDs)+len(params.Entry.Entries()))
		for i, uuid := range params.Entry.EntryUUIDs {
			hash, err := hex.DecodeString(uuid)
			if err != nil {
				return handleRekorAPIError(params, http.StatusBadRequest, err, malformedUUID)
			}
			searchHashes[i] = hash
		}

		code := http.StatusBadRequest
		for i, e := range params.Entry.Entries() {
			i, e := i, e // https://golang.org/doc/faq#closures_and_goroutines
			g.Go(func() error {
				entry, err := types.NewEntry(e)
				if err != nil {
					return err
				}
				if err := entry.Validate(); err != nil {
					return err
				}

				if entry.HasExternalEntities() {
					if err := entry.FetchExternalEntities(httpReqCtx); err != nil {
						return err
					}
				}

				leaf, err := entry.Canonicalize(httpReqCtx)
				if err != nil {
					code = http.StatusInternalServerError
					return err
				}
				hasher := rfc6962.DefaultHasher
				leafHash := hasher.HashLeaf(leaf)
				searchHashes[i+len(params.Entry.EntryUUIDs)] = leafHash
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return handleRekorAPIError(params, code, err, err.Error())
		}

		resp := tc.getLeafByHash(searchHashes) // TODO: if this API is deprecated, we need to ask for inclusion proof and then use index in proof result to get leaf
		switch resp.status {
		case codes.OK, codes.NotFound:
		default:
			return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianUnexpectedResult)
		}

		for _, leaf := range resp.getLeafResult.Leaves {
			logEntry := models.LogEntry{
				hex.EncodeToString(leaf.MerkleLeafHash): models.LogEntryAnon{
					LogIndex: &leaf.LeafIndex,
					Body:     leaf.LeafValue,
				},
			}
			resultPayload = append(resultPayload, logEntry)
		}
	}

	if len(params.Entry.LogIndexes) > 0 {
		g, _ := errgroup.WithContext(httpReqCtx)

		leaves := make([]*trillian.LogLeaf, len(params.Entry.LogIndexes))
		for i, logIndex := range params.Entry.LogIndexes {
			i, logIndex := i, logIndex // https://golang.org/doc/faq#closures_and_goroutines
			g.Go(func() error {
				resp := tc.getLeafByIndex(swag.Int64Value(logIndex))
				switch resp.status {
				case codes.OK, codes.NotFound:
				default:
					return resp.err
				}
				if resp.getLeafByRangeResult != nil {
					numLeaves := len(resp.getLeafByRangeResult.Leaves)
					if numLeaves == 0 {
						return nil
					} else if numLeaves != 1 {
						return errors.New("more than one leaf returned from getLeafByIndex call")
					}
					leaves[i] = resp.getLeafByRangeResult.Leaves[0]
				}
				return nil
			})
			if err := g.Wait(); err != nil {
				return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", err), trillianUnexpectedResult)
			}
		}

		for _, leaf := range leaves {
			if leaf != nil {
				logEntry := models.LogEntry{
					hex.EncodeToString(leaf.MerkleLeafHash): models.LogEntryAnon{
						LogIndex: &leaf.LeafIndex,
						Body:     leaf.LeafValue,
					},
				}
				resultPayload = append(resultPayload, logEntry)
			}
		}
	}

	return entries.NewSearchLogQueryOK().WithPayload(resultPayload)
}
