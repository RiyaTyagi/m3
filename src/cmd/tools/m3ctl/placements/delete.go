// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package placements

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/m3db/m3/src/cmd/tools/m3ctl/client"
)

// DoDelete does the delete api calls for placements
func DoDelete(
	endpoint string,
	service string,
	headers map[string]string,
	nodeName string,
	deleteEntire bool,
	logger *zap.Logger,
) ([]byte, error) {
	path := DefaultPath + service + "/placement"
	if deleteEntire {
		url := fmt.Sprintf("%s%s", endpoint, path)
		return client.DoDelete(url, headers, logger)
	}
	url := fmt.Sprintf("%s%s/%s", endpoint, path, nodeName)
	return client.DoDelete(url, headers, logger)
}
