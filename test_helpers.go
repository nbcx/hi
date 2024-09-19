// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package hi

import "net/http"

// CreateTestContext returns a fresh engine and context for testing purposes
func CreateTestContext(w http.ResponseWriter) (c *Context, r *Engine[*Context]) {
	r = New(&Context{})
	c = r.allocateContext(&Context{}, 0)
	c.Reset()
	c.writermem.reset(w)
	return
}

// CreateTestContextOnly returns a fresh context base on the engine for testing purposes
func CreateTestContextOnly(w http.ResponseWriter, r *Engine[*Context]) (c *Context) {
	c = r.allocateContext(&Context{}, r.maxParams)
	c.Reset()
	c.writermem.reset(w)
	return
}
