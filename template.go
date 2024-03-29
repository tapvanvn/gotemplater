package gotemplater

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/tapvanvn/gosmartstring"
	ss "github.com/tapvanvn/gosmartstring"
	"github.com/tapvanvn/gotemplater/tokenize/html"
	"github.com/tapvanvn/gotemplater/utility"
	"github.com/tapvanvn/gotokenize/v2"
)

type LanguageType = int

const (
	HTML LanguageType = iota
	JS
	CSS
	JSON
	TXT
	SS
)

type Template struct {
	ss.IObject

	ID           string //each teample has an id, the value is absolute path of that template file
	Path         []string
	uid          uuid.UUID
	Error        error
	IsReady      bool
	HostLanguage LanguageType
	Stream       gotokenize.TokenStream
	Context      *gosmartstring.SSContext
	instructions []*gotokenize.Token
}

func CreateTemplate(id string, hostLanguage LanguageType) Template {
	tpl := Template{
		IObject:      &ss.SSObject{},
		uid:          uuid.New(),
		ID:           id,
		Path:         []string{},
		Error:        nil,
		IsReady:      false,
		HostLanguage: hostLanguage,
		Stream:       gotokenize.CreateStream(0),
		Context:      gosmartstring.CreateContext(CreateHTMLRuntime()),
		instructions: []*gotokenize.Token{},
	}
	//tpl.Context.DebugLevel = 1
	return tpl
}

//GetRelativePath get
func (template *Template) GetRelativePath(path string) ([]string, error) {

	return utility.GetAbsolutePath(template.Path, path)
}

func (template *Template) load(isSmartstring bool) error {
	if template.Error != nil {

		return template.Error
	}
	path := "/" + strings.Join(template.Path, "/")

	//MARK: init route
	file, err := os.Open(path)

	if err != nil {
		//fmt.Println(err.Error())
		return err
	}

	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)

	stream := gotokenize.CreateStream(0)
	stream.Tokenize(string(bytes))

	if isSmartstring {

		template.Stream.Tokenize("{{" + string(bytes) + "}}")

	} else {

		template.Stream.Tokenize(string(bytes))
	}

	fmt.Println("load template from path:", path)

	return nil
}

func (template *Template) build(context *gosmartstring.SSContext) error {

	compiler := ss.SSCompiler{}
	template.Context.Reset()
	template.Context.BindingTo(context)

	if context.DebugLevel > 0 {
		template.Stream.Debug(0, html.HTMLTokenNaming, html.HTMLDebugOption)
		fmt.Println("--before build context--")
		context.PrintDebug(0)
		fmt.Println("----")
		template.Context.PrintDebug(0)
		fmt.Println("--end before build context--")
	}

	context.Reset()
	err := compiler.Compile(&template.Stream, template.Context)

	if context.DebugLevel > 1 {

		fmt.Println("--after build context--")
		context.PrintDebug(0)
		fmt.Println("----")
		template.Context.PrintDebug(0)
		fmt.Println("--end after build context--")
	}
	template.Context.BindingTo(nil)

	return err
}

func (template Template) CanExport() bool {
	return true
}

func (template Template) Export(context *gosmartstring.SSContext) []byte {

	var content = ""

	template.Context.BindingTo(context)
	template.Context.Reset()
	renderer := CreateRenderer()

	content, err := renderer.Compile(&template.Stream, template.Context)

	if err != nil {

		fmt.Println(err.Error())
	}

	template.Context.BindingTo(nil)

	return []byte(content)
}

func (obj Template) GetType() string {
	return "template->" + obj.ID
}
