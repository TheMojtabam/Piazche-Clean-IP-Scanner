package webui

import "io/fs"

// staticFS placeholder — در واقع embed میشه
// در اینجا برای compile فعلاً empty fs داریم
var staticFS fs.FS = emptyFS{}

type emptyFS struct{}
func (emptyFS) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }

// indexHTML — صفحه اصلی Web UI
// در build واقعی این با //go:embed embed میشه ولی برای سادگی inline هست
var indexHTML = []byte(indexHTMLContent)
