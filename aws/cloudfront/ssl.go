package cloudfront

//
// Minimum set of OpenSSL bindings to perform RSA-SHA1 signing
// using a PEM encoded RSA signature.
//

/*
#cgo pkg-config: libssl
#cgo darwin CFLAGS: -Wno-deprecated-declarations

#include <openssl/evp.h>
#include <openssl/ssl.h>
#include <openssl/bio.h>

int EVP_SignInit_not_a_macro(EVP_MD_CTX *ctx, const EVP_MD *type) { return EVP_SignInit(ctx, type); }
int EVP_SignUpdate_not_a_macro(EVP_MD_CTX *ctx, const void *d, unsigned int cnt) { return EVP_SignUpdate(ctx, d, cnt); }
*/
import "C"

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

var sslMutex = &sync.Mutex{}

type PrivateKey interface {
	// Signs the data using PKCS1.15
	SignPKCS1v15([]byte) ([]byte, error)
}

type pKey struct {
	key *C.EVP_PKEY
}

// LoadPrivateKeyFromPEM loads a private key from a PEM-encoded block.
func LoadPrivateKeyFromPEM(pem_block []byte) (PrivateKey, error) {
	//
	// Check and load the PEM data
	//
	if len(pem_block) == 0 {
		return nil, errors.New("empty pem block")
	}
	bio := C.BIO_new_mem_buf(unsafe.Pointer(&pem_block[0]),
		C.int(len(pem_block)))
	if bio == nil {
		return nil, errors.New("failed creating bio")
	}
	defer C.BIO_free(bio)

	rsakey := C.PEM_read_bio_RSAPrivateKey(bio, nil, nil, nil)
	if rsakey == nil {
		return nil, errors.New("failed reading rsa key")
	}
	defer C.RSA_free(rsakey)

	//
	// Create a private key
	//
	key := C.EVP_PKEY_new()
	if key == nil {
		return nil, errors.New("failed converting to evp_pkey")
	}
	if C.EVP_PKEY_set1_RSA(key, (*C.struct_rsa_st)(rsakey)) != 1 {
		C.EVP_PKEY_free(key)
		return nil, errors.New("failed converting to evp_pkey")
	}

	p := &pKey{key: key}
	runtime.SetFinalizer(p, func(p *pKey) {
		C.EVP_PKEY_free(p.key)
	})
	return p, nil
}

func (key *pKey) SignPKCS1v15(data []byte) ([]byte, error) {
	//
	// Initialize the context using sha1 as method, as we
	// need to generate RSA-SHA1 signature
	//
	var ctx C.EVP_MD_CTX
	C.EVP_MD_CTX_init(&ctx)
	defer C.EVP_MD_CTX_cleanup(&ctx)

	if 1 != C.EVP_SignInit_not_a_macro(&ctx, C.EVP_sha1()) {
		return nil, errors.New("rsasha1signature: failed to init signature")
	}
	if len(data) > 0 {
		if 1 != C.EVP_SignUpdate_not_a_macro(
			&ctx, unsafe.Pointer(&data[0]), C.uint(len(data))) {
			return nil, errors.New("rsasha1signature: failed to update signature")
		}
	}

	//
	// Sign
	//
	sig := make([]byte, C.EVP_PKEY_size(key.key))
	var sigblen C.uint

	// prevent data race when multiple threads are spawned
	// by the go runtime, mostly because you are performing the signing
	// inside a concurrent http request.
	sslMutex.Lock()
	defer sslMutex.Unlock()

	if 1 != C.EVP_SignFinal(&ctx,
		((*C.uchar)(unsafe.Pointer(&sig[0]))), &sigblen, key.key) {
		return nil, errors.New("rsasha1signature: failed to finalize signature")
	}
	return sig[:sigblen], nil
}
