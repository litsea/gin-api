package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type requestBody struct {
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

func newRequestBody(maxSize int, recordBody bool) *requestBody {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBuffer([]byte{})
	}

	return &requestBody{
		body:    body,
		maxSize: maxSize,
		bytes:   0,
	}
}

func (rb *requestBody) read(ctx *gin.Context) error {
	if rb.body == nil {
		return nil
	}

	if ctx.Request.Body == http.NoBody {
		return nil
	}

	var buf bytes.Buffer
	tee := io.TeeReader(ctx.Request.Body, &buf)
	body, err := io.ReadAll(tee)
	if err != nil {
		return fmt.Errorf("log.requestBody.read: %w", err)
	}

	ctx.Request.Body = io.NopCloser(&buf)
	rb.truncateUTF8(body)

	return err
}

func (rb *requestBody) truncateUTF8(b []byte) {
	if len(b) == 0 {
		return
	}

	if len(b) <= rb.maxSize {
		rb.bytes = len(b)
		rb.body = bytes.NewBuffer(b)
		return
	}

	i := 0
	for i < rb.maxSize {
		if i+utf8.RuneLen(rune(b[i])) > rb.maxSize {
			break
		}
		_, size := utf8.DecodeRune(b[i:])
		if i+size > rb.maxSize {
			break
		}
		i += size
	}

	rb.bytes = i
	rb.body = bytes.NewBuffer(b[:i])
}
