package opencc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rakyll/statik/fs"
	_ "github.com/tttlkkkl/opencc/statik"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var fsd http.FileSystem

func init() {
	var err error
	fsd, err = fs.New()
	if err != nil {
		panic("缺少字库文件，请使用statik命令重新打包")
	}
}

var punctuations []string = []string{
	" ", "\n", "\r", "\t", "-", ",", ".", "?", "!", "*", "　",
	"，", "。", "、", "；", "：", "？", "！", "…", "“", "”", "「",
	"」", "—", "－", "（", "）", "《", "》", "．", "／", "＼"}

type OpenCC struct {
	conf *Config
}

// Supported conversions: s2t, t2s, s2tw, tw2s, s2hk, hk2s, s2twp, tw2sp, t2tw, t2hk
func NewOpenCC(conversions string) (*OpenCC, error) {
	var err error

	file, err := fsd.Open(fmt.Sprintf("/config/%s.json", conversions))
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var conf *Config
	err = json.Unmarshal(body, &conf)
	if err != nil {
		return nil, err
	}
	err = conf.init()
	if err != nil {
		return nil, err
	}
	//
	return &OpenCC{conf: conf}, nil
}

//
func (oc *OpenCC) ConvertFile(in io.Reader, out io.Writer) error {
	inReader := bufio.NewReader(in)
	for {
		lineText, readErr := inReader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		nLineText, err := oc.splitText(lineText)
		if err != nil {
			return err
		}
		_, err = out.Write([]byte(nLineText))
		if err != nil {
			return err
		}
		if readErr == io.EOF {
			break
		}
	}
	return nil
}

//
func (oc *OpenCC) ConvertText(text string) (string, error) {
	return oc.splitText(text)
}

//
func (oc *OpenCC) splitText(text string) (string, error) {
	tmp := make([]string, 0, len(text))
	var newText string
	for i, c := range strings.Split(text, "") {
		if i > 0 && isPunctuations(c) {
			if len(tmp) > 0 {
				tx, err := oc.convertString(strings.Join(tmp, ""))
				if err != nil {
					return text, err
				}
				newText = newText + tx + c
				tmp = tmp[:0]
			} else {
				newText = newText + c
			}
			continue
		}
		tmp = append(tmp, c)
	}
	if len(tmp) > 0 {
		tx, err := oc.convertString(strings.Join(tmp, ""))
		if err != nil {
			return text, err
		}
		newText = newText + tx
	}
	return newText, nil
}

//
func (oc *OpenCC) convertString(text string) (string, error) {
	var err error
	if oc.conf == nil {
		return text, fmt.Errorf("no config")
	}
	text, err = oc.conf.convertText(text)
	if err != nil {
		return text, err
	}
	return text, nil
}

//是否标点符号
func isPunctuations(character string) bool {
	if len([]byte(character)) <= 1 {
		return true
	}
	//
	for _, c := range punctuations {
		if c == character {
			return true
		}
	}
	return false
}
