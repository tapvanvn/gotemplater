package gotemplater

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tapvanvn/gosmartstring"
	ss "github.com/tapvanvn/gosmartstring"
	"github.com/tapvanvn/gotemplater/tokenize/html"
	"github.com/tapvanvn/gotemplater/utility"
	"github.com/tapvanvn/gotokenize/v2"
)

var __templater *templater = nil
var __htmlMeaning *html.HTMLOptmizerMeaning = nil
var __ssMeaning *gosmartstring.SmarstringInstructionMeaning = nil

func init() {
	gosmartstring.SSInsructionMove(5000)
	__ssMeaning = gosmartstring.CreateSSInstructionMeaning()
	__htmlMeaning = html.CreateHTMLOptmizer()
	__templater = &templater{
		namespaces:     map[string][]string{},
		loadedTemplate: map[string]*Template{},
	}
}

//Templater manage
type templater struct {
	namespaces     map[string][]string //namespace to path
	loadedTemplate map[string]*Template
}

func GetTemplater() *templater {

	return __templater
}

//MARK: implement functions

func (tpt *templater) Debug() {

	fmt.Println("Namespaces")
	for namespace, path := range tpt.namespaces {
		fmt.Println(namespace, ":", path)
	}
}

//AddNamespace add a namespace
func (tpt *templater) AddNamespace(namespace string, path string) error {

	segments := strings.Split(path, "/")
	resultSegments := []string{}
	for _, segment := range segments {
		if segment == "." {

		} else if segment == ".." {
			numSegment := len(resultSegments)
			if numSegment > 0 {
				resultSegments = resultSegments[0 : numSegment-1]
			} else {
				return errors.New("path error")
			}
		} else if len(segment) > 0 {

			resultSegments = append(resultSegments, segment)
		}
	}
	tpt.namespaces[namespace] = resultSegments
	return nil
}

func (tpt *templater) GetPath(id string) ([]string, error) {
	relativePath := strings.TrimSpace(id)
	sep := strings.Index(relativePath, ":")

	namespace := ""
	if sep >= 0 {
		namespace = relativePath[0:sep]
		relativePath = relativePath[sep+1:]
	}
	nsPathSegments, ok := tpt.namespaces[namespace]

	if !ok {

		return nil, errors.New("namespace is not defined")
	}

	return utility.GetAbsolutePath(nsPathSegments, relativePath)
}

func (tpt *templater) Render(id string, context *gosmartstring.SSContext) (string, error) {

	instructionDo := ss.BuildDo("template",
		[]ss.IObject{ss.CreateString(id)}, context)

	stream := gotokenize.CreateStream(0)
	stream.AddToken(instructionDo)
	compiler := ss.SSCompiler{}
	if err := compiler.Compile(&stream, context); err != nil {
		fmt.Println(err.Error())
		context.PrintDebug(0)
		return "", err
	}
	//return "", nil
	renderer := CreateRenderer()
	return renderer.Compile(&stream, context)
}

func (tpt *templater) ClearAllCache() {

	tpt.loadedTemplate = map[string]*Template{}
}

func (tpt *templater) ClearCache(id string) {

	delete(tpt.loadedTemplate, id)
}

func (tpt *templater) GetTemplate(id string) *Template {

	if template, ok := tpt.loadedTemplate[id]; ok {
		//fmt.Println("template loaded:", id)
		return template
	}
	template := CreateTemplate(id, TXT)

	tpt.loadedTemplate[id] = &template

	loadPath, err := tpt.GetPath(id)
	if err != nil {
		fmt.Println(err.Error())
		template.Error = err
		return &template
	}
	numSegment := len(loadPath)

	lastSegment := loadPath[numSegment-1]

	extPos := strings.LastIndex(lastSegment, ".")
	var languageProcessor gotokenize.IMeaning = nil
	if extPos >= 0 {
		ext := strings.ToLower(lastSegment[extPos+1:])
		if ext == "html" || ext == "htm" {
			template.HostLanguage = HTML
			languageProcessor = __htmlMeaning
		} else if ext == "js" {
			template.HostLanguage = JS
		} else if ext == "json" {
			template.HostLanguage = JSON
		} else if ext == "ss" {
			template.HostLanguage = SS
			languageProcessor = __ssMeaning
		}
	}
	template.Path = loadPath

	err = template.load(template.HostLanguage == SS)

	if err != nil {
		template.Error = err
		fmt.Println(err.Error())
		return &template
	}

	proc := gotokenize.NewMeaningProcessFromStream(gotokenize.NoTokens, &template.Stream)
	proc.Context.BindingData = template.Context

	if languageProcessor != nil {
		languageProcessor.Prepare(proc)

		tmpStream := gotokenize.CreateStream(0)
		for {
			token := languageProcessor.Next(proc)
			if token == nil {
				break
			}
			tmpStream.AddToken(*token)
		}
		template.Stream = tmpStream
	}

	// fmt.Println("--build--")
	// tmpStream.Debug(0, html.HTMLTokenNaming, html.HTMLDebugOption)
	// fmt.Println("--end build--")

	template.IsReady = true

	return &template
}
