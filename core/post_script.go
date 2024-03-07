package core

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type PostScriptConfig struct {
	Func string   `json:"func"`
	Src  []string `json:"src"`
	Dst  []string `json:"dst"`
}

const (
	FuncNameStringToJson string = "stringToJSON"
)

var FuncStringToJson = func(in string, src, dst []string) string {

	vStr := gjson.Get(in, strings.Join(src, "."))

	resStr, _ := sjson.SetRaw(in, strings.Join(dst, "."), vStr.String())

	var resBuf bytes.Buffer
	json.Indent(&resBuf, []byte(resStr), "", "  ")

	return resBuf.String()
}
