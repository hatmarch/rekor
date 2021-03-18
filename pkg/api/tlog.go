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
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/sigstore/rekor/pkg/generated/models"
	"google.golang.org/grpc/codes"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	tclient "github.com/google/trillian/client"
	tcrypto "github.com/google/trillian/crypto"
	rfc6962 "github.com/google/trillian/merkle/rfc6962/hasher"
	"github.com/sigstore/rekor/pkg/generated/restapi/operations/tlog"
)

func GetLogInfoHandler(params tlog.GetLogInfoParams) middleware.Responder {
	tc := NewTrillianClient(params.HTTPRequest.Context())

	resp := tc.getLatest(0)
	if resp.status != codes.OK {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianCommunicationError)
	}
	result := resp.getLatestResult

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

	hashString := hex.EncodeToString(root.RootHash)
	treeSize := int64(root.TreeSize)
	keyHint := strfmt.Base64(result.SignedLogRoot.GetKeyHint())
	logRoot := strfmt.Base64(result.SignedLogRoot.GetLogRoot())
	signature := strfmt.Base64(result.SignedLogRoot.GetLogRootSignature())

	sth := models.LogInfoSignedTreeHead{
		KeyHint:   &keyHint,
		LogRoot:   &logRoot,
		Signature: &signature,
	}

	logInfo := models.LogInfo{
		RootHash:       &hashString,
		TreeSize:       &treeSize,
		SignedTreeHead: &sth,
	}
	return tlog.NewGetLogInfoOK().WithPayload(&logInfo)
}

func GetLogProofHandler(params tlog.GetLogProofParams) middleware.Responder {
	if *params.FirstSize > params.LastSize {
		return handleRekorAPIError(params, http.StatusBadRequest, nil, fmt.Sprintf(firstSizeLessThanLastSize, *params.FirstSize, params.LastSize))
	}
	tc := NewTrillianClient(params.HTTPRequest.Context())

	resp := tc.getConsistencyProof(*params.FirstSize, params.LastSize)
	if resp.status != codes.OK {
		return handleRekorAPIError(params, http.StatusInternalServerError, fmt.Errorf("grpc error: %w", resp.err), trillianCommunicationError)
	}
	result := resp.getConsistencyProofResult

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

	hashString := hex.EncodeToString(root.RootHash)
	proofHashes := []string{}

	if proof := result.GetProof(); proof != nil {
		for _, hash := range proof.Hashes {
			proofHashes = append(proofHashes, hex.EncodeToString(hash))
		}
	} else {
		// The proof field may be empty if the requested tree_size was larger than that available at the server
		// (e.g. because there is skew between server instances, and an earlier client request was processed by
		// a more up-to-date instance). root.TreeSize is the maximum size currently observed
		return handleRekorAPIError(params, http.StatusBadRequest, nil, fmt.Sprintf(lastSizeGreaterThanKnown, params.LastSize, root.TreeSize))
	}

	consistencyProof := models.ConsistencyProof{
		RootHash: &hashString,
		Hashes:   proofHashes,
	}
	return tlog.NewGetLogProofOK().WithPayload(&consistencyProof)
}

func GetPublicKeyHandler(params tlog.GetPublicKeyParams) middleware.Responder {
	tc := NewTrillianClient(params.HTTPRequest.Context())

	keyBuf := bytes.Buffer{}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: tc.pubkey.Der,
	}

	if err := pem.Encode(&keyBuf, block); err != nil {
		return handleRekorAPIError(params, http.StatusInternalServerError, err, trillianUnexpectedResult)
	}

	return tlog.NewGetPublicKeyOK().WithPayload(keyBuf.String())
}
