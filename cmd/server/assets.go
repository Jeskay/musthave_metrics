package main

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assetsea37ae207d5121d1af028163ef30a53b86588dd5 = "<html>\r\n    <h1>\r\n        {{ .title }}\r\n    </h1>\r\n    <ol>\r\n    {{ range .Metrics }}\r\n        <li>\r\n        Name: {{.Name}}<br>\r\n        Type: {{.Type}}<br>\r\n        Value: {{.Value}}<br>\r\n        </li>\r\n    {{ end }}\r\n    </ol>\r\n</html>"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"templates"}, "/templates": []string{"list.tmpl"}}, map[string]*assets.File{
	"/templates": &assets.File{
		Path:     "/templates",
		FileMode: 0x800001ff,
		Mtime:    time.Unix(1728164562, 1728164562699619700),
		Data:     nil,
	}, "/templates/list.tmpl": &assets.File{
		Path:     "/templates/list.tmpl",
		FileMode: 0x1b6,
		Mtime:    time.Unix(1728214737, 1728214737567687000),
		Data:     []byte(_Assetsea37ae207d5121d1af028163ef30a53b86588dd5),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ff,
		Mtime:    time.Unix(1728213996, 1728213996192661100),
		Data:     nil,
	}}, "")
